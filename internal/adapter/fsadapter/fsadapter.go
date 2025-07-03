package fsadapter

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/jgivc/fetchtracker/internal/entity"
)

const (
	maxFiles        = 100
	mimeTypeUnknown = "application/octet-stream"
)

type fileDesc struct {
	id       string
	fileName string
	filePath string
	fileSize int64
	mimeType string
}

type fsAdapter struct {
	descFileName string
	skipFiles    map[string]struct{}
	log          *slog.Logger
}

func (a *fsAdapter) ToDownload(folderPath string) (*entity.Download, error) {
	panic("not implemented")
}

func (a *fsAdapter) readFiles(folderPath string) ([]*fileDesc, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var files []*fileDesc
	for _, entry := range entries {
		if !entry.IsDir() {
			fDesc := &fileDesc{
				fileName: entry.Name(),
				filePath: path.Join(folderPath, entry.Name()),
			}

			if _, exists := a.skipFiles[fDesc.fileName]; exists {
				a.log.Info("Skip file", slog.String("path", fDesc.filePath))

				continue
			}

			fDesc.id = getIDFromPath(fDesc.filePath)

			stat, err := os.Stat(fDesc.filePath)
			if err != nil {
				a.log.Error("Cannot get file size", slog.String("path", fDesc.filePath), slog.Any("error", err))
			} else {
				fDesc.fileSize = stat.Size()
			}

			mimeType, err := getMimeType(fDesc.filePath)
			if err != nil {
				a.log.Error("Cannot get file mimeType", slog.String("path", fDesc.filePath), slog.Any("error", err))
			} else {
				fDesc.mimeType = mimeType
			}

			files = append(files, fDesc)
		}

		if len(files) >= maxFiles {
			break
		}
	}

	return files, nil
}

func getMimeType(filePath string) (string, error) {
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
