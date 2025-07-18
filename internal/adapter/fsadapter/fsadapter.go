package fsadapter

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "embed"

	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/jgivc/fetchtracker/internal/util"
	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

const (
	maxFiles        = 100
	mimeTypeUnknown = "application/octet-stream"
	templateName    = "tmpl"
)

//go:embed template.html
var defaultTemplate string

var filesPlural = []string{"файл", "файла", "файлов"}

type FolderDesc struct {
	Title   string            `yaml:"title"`
	Enabled bool              `yaml:"enabled"`
	Files   map[string]string `yaml:"files"`
}

type fsAdapter struct {
	fs           afero.Fs
	rootDor      string
	descFileName string
	url          string
	skipFiles    map[string]struct{}
	md           goldmark.Markdown
	tmpl         *template.Template
	log          *slog.Logger
}

func NewFSAdapter(rootDir string, descFileName string, tmplFileName string, url string, skipFiles []string, log *slog.Logger) (*fsAdapter, error) {
	return NewFSAdapterWithFS(afero.NewOsFs(), rootDir, descFileName, tmplFileName, skipFiles, log)
}

func NewFSAdapterWithFS(fs afero.Fs, rootDir string, descFileName string, tmplFileName string, skipFiles []string, log *slog.Logger) (*fsAdapter, error) {
	skipFilesMap := make(map[string]struct{})
	skipFilesMap[descFileName] = struct{}{}
	for _, file := range skipFiles {
		skipFilesMap[file] = struct{}{}
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
		),
	)

	fsa := &fsAdapter{
		fs:           fs,
		rootDor:      rootDir,
		descFileName: descFileName,
		skipFiles:    skipFilesMap,
		md:           md,
		log:          log,
	}

	var (
		tmpl *template.Template
		err  error
	)

	if tmplFileName == "" {
		tmpl, err = template.New(templateName).Funcs(template.FuncMap{"plural": plural}).Parse(defaultTemplate)
	} else {
		tmpl, err = template.New(templateName).ParseFiles(tmplFileName)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot parse template: %w", err)
	}

	fsa.tmpl = tmpl

	return fsa, nil
}

func (a *fsAdapter) ToDownload(folderPath string) (*entity.Download, error) {
	if strings.Contains(folderPath, "..") {
		return nil, fmt.Errorf("invalid folder path")
	}

	files, err := a.readFiles(folderPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get folder files: %w", err)
	}

	if len(files) < 1 {
		return nil, fmt.Errorf("folder have no files")
	}

	descr, fd, err := a.getFolderDesc(folderPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get folder description: %w", err)
	}

	download := &entity.Download{
		ID:          util.GetIDFromString(&folderPath),
		Title:       fd.Title,
		PageContent: descr,
		Enabled:     fd.Enabled,
		SourcePath:  folderPath,
		CreatedAt:   time.Now(),
	}

	if len(fd.Files) > 0 {
		for i := range files {
			if fileDesc, exists := fd.Files[files[i].Name]; exists {
				files[i].Description = fileDesc
			}
		}
	}

	download.Files = files

	var buf bytes.Buffer
	download.PageContent = descr
	if err := a.tmpl.Execute(&buf, download); err != nil {
		return nil, fmt.Errorf("cannot build download page: %w", err)
	}

	download.PageContent = buf.String()
	download.PageHash = util.GetIDFromString(&download.PageContent)
	// fmt.Println(download.PageContent)

	return download, nil
}

func (a *fsAdapter) readFiles(folderPath string) ([]*entity.File, error) {
	entries, err := afero.ReadDir(a.fs, folderPath)
	if err != nil {
		return nil, err
	}

	var files []*entity.File
	for _, entry := range entries {
		if !entry.IsDir() {
			fDesc := &entity.File{
				Name:       entry.Name(),
				SourcePath: filepath.Join(folderPath, entry.Name()),
			}

			fDesc.URL = filepath.Join("/", filepath.Base(a.rootDor), strings.Replace(fDesc.SourcePath, a.rootDor, "/", 1))
			fmt.Println("AAA", fDesc)

			if _, exists := a.skipFiles[fDesc.Name]; exists {
				a.log.Info("Skip file", slog.String("path", fDesc.SourcePath))

				continue
			}

			fDesc.ID = util.GetIDFromString(&fDesc.SourcePath)

			stat, err := a.fs.Stat(fDesc.SourcePath)
			if err != nil {
				a.log.Error("Cannot get file size", slog.String("path", fDesc.SourcePath), slog.Any("error", err))
			} else {
				fDesc.Size = stat.Size()
			}

			mimeType, err := a.getMimeType(fDesc.SourcePath)
			if err != nil {
				a.log.Error("Cannot get file mimeType", slog.String("path", fDesc.SourcePath), slog.Any("error", err))
			} else {
				fDesc.MIMEType = mimeType
			}

			files = append(files, fDesc)
		}

		if len(files) >= maxFiles {
			break
		}
	}

	return files, nil
}

func (a *fsAdapter) getFolderDesc(folderPath string) (string, *FolderDesc, error) {
	filePath := filepath.Join(folderPath, a.descFileName)
	if !a.fileExists(filePath) {
		return "", &FolderDesc{
			Title:   filepath.Base(folderPath),
			Enabled: true,
		}, nil
	}

	data, err := afero.ReadFile(a.fs, filePath)
	if err != nil {
		return "", nil, fmt.Errorf("cannot read description file")
	}

	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := a.md.Convert(data, &buf, parser.WithContext(ctx)); err != nil {
		return "", nil, fmt.Errorf("cannot convert description file to markdown")
	}

	fm := frontmatter.Get(ctx)
	var fd FolderDesc
	if err := fm.Decode(&fd); err != nil {
		return "", nil, fmt.Errorf("cannot read frontmatter from description file")
	}

	return buf.String(), &fd, nil
}

func (a *fsAdapter) getMimeType(filePath string) (string, error) {
	if ext := filepath.Ext(filePath); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType, nil
		}
	}

	file, err := a.fs.Open(filePath)
	if err != nil {
		return mimeTypeUnknown, err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return mimeTypeUnknown, err
	}

	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func (a *fsAdapter) fileExists(path string) bool {
	_, err := a.fs.Stat(path)
	if err == nil {
		return true // Path exists
	}

	if os.IsNotExist(err) {
		return false
	}
	// Other errors (e.g., permission issues)
	// fmt.Printf("Error checking path %s: %v\n", path, err)
	return false
}

func plural(n int) string {
	var idx int
	// @see http://docs.translatehouse.org/projects/localization-guide/en/latest/l10n/pluralforms.html

	switch {
	case n%10 == 1 && n%100 != 11:
	case (n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20)):
		idx = 1
	default:
		idx = 2
	}

	if idx >= len(filesPlural) {
		return ""
	}

	return filesPlural[idx]
}

// func (a *fsAdapter) buildDownloadPage(desc string, download *entity.Download) (string, error) {
// 	panic("not implemented")
// }
