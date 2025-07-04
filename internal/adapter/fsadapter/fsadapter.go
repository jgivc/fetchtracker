package fsadapter

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

const (
	maxFiles        = 100
	mimeTypeUnknown = "application/octet-stream"
)

type FolderDesc struct {
	Title   string            `yaml:"title"`
	Enabled bool              `yaml:"enabled"`
	Files   map[string]string `yaml:"files"`
}

type fsAdapter struct {
	fs           afero.Fs
	descFileName string
	skipFiles    map[string]struct{}
	md           goldmark.Markdown
	log          *slog.Logger
}

// Добавьте конструктор
func NewFSAdapter(descFileName string, skipFiles []string, log *slog.Logger) *fsAdapter {
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

	return &fsAdapter{
		descFileName: descFileName,
		skipFiles:    skipFilesMap,
		md:           md,
		log:          log,
	}
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
		ID:          getIDFromPath(folderPath),
		Title:       fd.Title,
		Description: descr,
		Enabled:     fd.Enabled,
		SourcePath:  folderPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if len(fd.Files) > 0 {
		for i := range files {
			if fileDesc, exists := fd.Files[files[i].Name]; exists {
				files[i].Description = fileDesc
			}
		}
	}

	download.Files = files

	return download, nil
}

func (a *fsAdapter) readFiles(folderPath string) ([]*entity.File, error) {
	entries, err := os.ReadDir(folderPath)
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

			if _, exists := a.skipFiles[fDesc.Name]; exists {
				a.log.Info("Skip file", slog.String("path", fDesc.SourcePath))

				continue
			}

			fDesc.ID = getIDFromPath(fDesc.SourcePath)

			stat, err := os.Stat(fDesc.SourcePath)
			if err != nil {
				a.log.Error("Cannot get file size", slog.String("path", fDesc.SourcePath), slog.Any("error", err))
			} else {
				fDesc.Size = stat.Size()
			}

			mimeType, err := getMimeType(fDesc.SourcePath)
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
	if !fileExists(filePath) {
		return "", &FolderDesc{
			Title:   filepath.Base(folderPath),
			Enabled: true,
		}, nil
	}

	data, err := os.ReadFile(filePath)
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

func getMimeType(filePath string) (string, error) {
	if ext := filepath.Ext(filePath); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType, nil
		}
	}

	file, err := os.Open(filePath)
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

func getIDFromPath(filePath string) string {
	hasher := sha1.New()
	hasher.Write([]byte(filePath))

	return hex.EncodeToString(hasher.Sum(nil))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // Path exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false // Path does not exist
	}
	// Other errors (e.g., permission issues)
	// fmt.Printf("Error checking path %s: %v\n", path, err)
	return false
}
