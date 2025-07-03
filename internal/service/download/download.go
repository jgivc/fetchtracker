package download

import (
	"fmt"
	"log/slog"
)

const (
	serviceName = "download"
)

type DownloadRepository interface {
	GetFilePath(id string) (string, error)
	IncFileCounter(id string) error
}

type downloadService struct {
	repo DownloadRepository
	log  *slog.Logger
}

func NewDownloadService(repo DownloadRepository, log *slog.Logger) *downloadService {
	return &downloadService{
		repo: repo,
		log:  log.With(slog.String("service", serviceName)),
	}
}

func (d *downloadService) Download(id string) (string, error) {
	filePath, err := d.repo.GetFilePath(id)
	if err != nil {
		d.log.Error("Cannot get file path", slog.String("file_id", id), slog.Any("error", err))

		return "", fmt.Errorf("cannot get file path: %w", err)
	}

	if err := d.repo.IncFileCounter(id); err != nil {
		d.log.Error("Cannot increment file counter", slog.String("file_id", id), slog.Any("error", err))
	}

	return filePath, nil
}
