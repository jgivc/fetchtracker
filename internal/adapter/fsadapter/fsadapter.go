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

	"github.com/jgivc/fetchtracker/internal/adapter/fsadapter/mdadapter"
	"github.com/jgivc/fetchtracker/internal/config"
	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/jgivc/fetchtracker/internal/util"
	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

const (
	ParseModeIndex ParseMode = iota
	ParseModeDefaultIndex
	ParseModeMdCustomTemplate
	ParseModeMdDefaultTemplate

	maxFiles              = 100
	mimeTypeUnknown       = "application/octet-stream"
	mimeTypeCheckPartSize = 512

	templateNameFile  = "FILE"
	templateNameFiles = "FILES"

	funcNameFile  = "file"
	funcNameFiles = "files"
)

type ParseMode int

func (m ParseMode) String() string {
	return [...]string{"Index", "DefaultIndex", "MDCustom", "MDDefault"}[m]
}

var (
	//go:embed templates/template.html
	defaultTemplateContent []byte

	//go:embed templates/index.html
	defaultIndexContent []byte
)

type PageContextIndex struct {
	URL string
	*entity.Download
}

type PageContext struct {
	URL         string
	ContentHTML template.HTML
	*entity.Download
	Frontmatter *Frontmatter
}

type Frontmatter struct {
	Title   string            `yaml:"title"`
	Enabled bool              `yaml:"enabled"`
	Files   map[string]string `yaml:"files"`
	Author  string            `yaml:"author"`
}

type fsAdapter struct {
	fs        afero.Fs
	cfg       *config.FSAdapterConfig
	skipFiles map[string]struct{}
	md        goldmark.Markdown

	log *slog.Logger
}

func NewFSAdapter(cfg *config.FSAdapterConfig, log *slog.Logger) (*fsAdapter, error) {
	return NewFSAdapterWithFS(afero.NewOsFs(), cfg, log)
}

func NewFSAdapterWithFS(fs afero.Fs, cfg *config.FSAdapterConfig, log *slog.Logger) (*fsAdapter, error) {
	skipFilesMap := make(map[string]struct{})
	skipFilesMap[cfg.IndexPageFileName] = struct{}{}
	skipFilesMap[cfg.DescFileName] = struct{}{}
	skipFilesMap[cfg.TemplateFileName] = struct{}{}
	for _, file := range cfg.SkipFiles {
		skipFilesMap[file] = struct{}{}
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			mdadapter.NewFilesExtension(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	fsa := &fsAdapter{
		fs:        fs,
		cfg:       cfg,
		skipFiles: skipFilesMap,
		md:        md,
		log:       log,
	}

	return fsa, nil
}

/*
1. If the distribution folder contains the file cfg.IndexPageFileName, then parse it.
2. If the distribution folder does not contain the file cfg.DescFileName, then parse the default cfg.DefaultIndexTemplate template.
3. If the distribution folder contains cfg.TemplateFileName and cfg.DescFileName, then parse it with it.
4. If the folder only contains cfg.DescFileName, then parse with cfg.DefaultMDTemplate template.
*/
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

	download := &entity.Download{
		ID:         util.GetIDFromString(&folderPath),
		Title:      filepath.Base(folderPath),
		Enabled:    true,
		SourcePath: folderPath,
		CreatedAt:  time.Now(),
		Files:      files,
	}

	switch a.getParseMode(folderPath) {
	case ParseModeIndex, ParseModeDefaultIndex:
		if err := a.parseIndex(folderPath, download); err != nil {
			return nil, fmt.Errorf("cannot parse index page: %w", err)
		}
	case ParseModeMdCustomTemplate, ParseModeMdDefaultTemplate:
		if err := a.parseMarkdown(folderPath, download); err != nil {
			return nil, fmt.Errorf("cannot parse index page: %w", err)
		}
	}

	return download, nil
}

func (a *fsAdapter) parseIndex(folderPath string, download *entity.Download) error {
	tmpl, err := a.getTemplate(filepath.Join(folderPath, a.cfg.IndexPageFileName), defaultIndexContent, nil, download.Files)
	if err != nil {
		return fmt.Errorf("cannot get index template: %w", err)
	}

	content, err := buildTemplate(tmpl, &PageContextIndex{URL: a.cfg.URL, Download: download})
	if err != nil {
		return fmt.Errorf("cannot build index template: %w", err)
	}

	download.PageContent = content
	download.PageHash = util.GetIDFromString(&content)

	return nil
}

func (a *fsAdapter) parseMarkdown(folderPath string, download *entity.Download) error {
	mdFileName := filepath.Join(folderPath, a.cfg.DescFileName)

	fm, mdData, err := a.getFrontmatter(mdFileName)
	if err != nil {
		return fmt.Errorf("cannot get frontmatter: %w", err)
	}

	if fm != nil {
		download.Title = fm.Title
		download.Enabled = fm.Enabled

		if len(fm.Files) > 0 {
			for i := range download.Files {
				if fileDesc, exists := fm.Files[download.Files[i].Name]; exists {
					download.Files[i].Description = fileDesc
				}
			}
		}
	}

	// Load template user or my own
	tmpl, err := a.getTemplate(filepath.Join(folderPath, a.cfg.TemplateFileName), defaultTemplateContent, nil, nil)
	if err != nil {
		return fmt.Errorf("cannot get page template: %w", err)
	}

	// Get templateResolver
	tResolver, err := newTemplateResolver(tmpl)
	if err != nil {
		// Since it is not known which template was loaded, the user's or ours,
		// we explicitly load our template to search for predefined templates FILE and FILES
		ftmpl, err := a.getTemplate("", defaultTemplateContent, nil, nil)
		if err != nil {
			return fmt.Errorf("cannot get own template: %w", err)
		}

		tResolver, err = newTemplateResolver(ftmpl)
		if err != nil {
			return fmt.Errorf("cannot get template: %w", err)
		}
	}

	pc := parser.NewContext()
	pc.Set(mdadapter.TemplateResolverKey, tResolver)
	pc.Set(mdadapter.FileResolverKey, newFileResolver(download.Files))

	// Convert markdown to html
	var buf bytes.Buffer
	if err := a.md.Convert(mdData, &buf, parser.WithContext(pc)); err != nil {
		return fmt.Errorf("cannot convert markdown: %w", err)
	}

	// Convert entire page
	content, err := buildTemplateHTML(tmpl, &PageContext{URL: a.cfg.URL, ContentHTML: template.HTML(buf.String()), Download: download, Frontmatter: fm})
	if err != nil {
		return fmt.Errorf("cannot build page: %w", err)
	}

	download.PageContent = string(content)
	download.PageHash = util.GetIDFromString(&download.PageContent)

	return nil
}

func (a *fsAdapter) getParseMode(folderPath string) ParseMode {
	if indexFileName := filepath.Join(folderPath, a.cfg.IndexPageFileName); a.fileExists(indexFileName) {
		return ParseModeIndex
	}

	if mdFileName := filepath.Join(folderPath, a.cfg.DescFileName); !a.fileExists(mdFileName) {
		return ParseModeDefaultIndex
	}

	if customTemplateFileName := filepath.Join(folderPath, a.cfg.TemplateFileName); a.fileExists(customTemplateFileName) {
		return ParseModeMdCustomTemplate
	}

	return ParseModeMdDefaultTemplate
}

func (a *fsAdapter) getFrontmatter(fileName string) (*Frontmatter, []byte, error) {
	content, err := afero.ReadFile(a.fs, fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read md file: %w", err)
	}

	var fm Frontmatter
	str := string(content)

	if !strings.HasPrefix(str, "---\n") {
		return nil, []byte(str), nil
	}

	parts := strings.SplitN(str, "---\n", 3)
	if len(parts) < 3 {
		return nil, []byte(str), nil
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, nil, fmt.Errorf("cannot unmarshal frontmatter: %w", err)
	}

	return &fm, []byte(parts[2]), nil
}

func (a *fsAdapter) getTemplate(templateFileName string, defaultTemplateContent []byte, parentTmpl *template.Template, files []*entity.File) (*template.Template, error) {
	var (
		content []byte
		err     error
	)

	if a.fileExists(templateFileName) {
		content, err = afero.ReadFile(a.fs, templateFileName)
		if err != nil {
			return nil, fmt.Errorf("cannot read template file: %s: %w", templateFileName, err)
		}
	} else {
		content = defaultTemplateContent
	}

	if content == nil {
		return nil, fmt.Errorf("cannot get template content")
	}

	tmpl := template.New("")
	if parentTmpl != nil {
		tmpl = parentTmpl
	}
	if len(files) > 0 {
		filesMap := make(map[string]*entity.File, len(files))
		for i := range files {
			filesMap[files[i].Name] = files[i]
		}

		funcMap := template.FuncMap{}
		if tmpl.Lookup(templateNameFile) != nil {
			funcMap[funcNameFile] = func(fileName string, args ...string) (template.HTML, error) {
				tt := tmpl.Lookup(templateNameFile)
				if tt == nil {
					return "", fmt.Errorf("template %s must be defined", templateNameFile)
				}

				file, exists := filesMap[fileName]
				if !exists {
					return "", fmt.Errorf("cannot find file: %s", fileName)
				}

				if len(args) > 0 {
					file.Description = args[0]
				}

				return buildTemplateHTML(tmpl, file)
			}
		}

		if tmpl.Lookup(templateNameFiles) != nil {
			funcMap[funcNameFiles] = func() (template.HTML, error) {
				tt := tmpl.Lookup(templateNameFiles)
				if tt == nil {
					return "", fmt.Errorf("template %s must be defined", templateNameFiles)
				}

				return buildTemplateHTML(tt, files)
			}
		}

		if len(funcMap) > 0 {
			tmpl = tmpl.Funcs(funcMap)

		}
	}

	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("cannot parse template content: %w", err)
	}

	return tmpl, nil
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

			fDesc.URL = filepath.Join("/", filepath.Base(a.cfg.WorkDir), strings.Replace(fDesc.SourcePath, a.cfg.WorkDir, "/", 1))

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

func buildTemplate(tmpl *template.Template, data any) (string, error) {
	buf := bytes.Buffer{}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("cannot execute template: %w", err)
	}

	return buf.String(), nil
}

func buildTemplateHTML(tmpl *template.Template, data any) (template.HTML, error) {
	content, err := buildTemplate(tmpl, data)

	return template.HTML(content), err
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

	buffer := make([]byte, mimeTypeCheckPartSize)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return mimeTypeUnknown, err
	}

	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func (a *fsAdapter) fileExists(path string) bool {
	if path == "" {
		return false
	}

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
