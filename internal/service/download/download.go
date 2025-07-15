package download

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	serviceName = "download"
)

type DownloadRepository interface {
	GetFilePath(ctx context.Context, id string) (string, error)
	UserExists(ctx context.Context, id string) (bool, error)
	IncFileCounter(ctx context.Context, id string) (int64, error)
	GetPage(ctx context.Context, id string) (string, error)
	GetDownloadCounters(ctx context.Context, id string) (map[string]int, error)
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

func (d *downloadService) Download(ctx context.Context, id string) (string, error) {
	filePath, err := d.repo.GetFilePath(ctx, id)
	if err != nil {
		d.log.Error("Cannot get file path", slog.String("file_id", id), slog.Any("error", err))

		return "", fmt.Errorf("cannot get file path: %w", err)
	}

	return filePath, nil
}

func (d *downloadService) IncFileCounter(ctx context.Context, userID, fileID string) (int64, error) {

	exists, err := d.repo.UserExists(ctx, userID)
	if err != nil {
		d.log.Error("Cannot check user exists", slog.String("user_id", userID), slog.String("file_id", fileID), slog.Any("error", err))

		return 0, fmt.Errorf("cannot check user exists: %w", err)
	}

	if !exists {
		counter, err := d.repo.IncFileCounter(ctx, fileID)
		if err != nil {
			d.log.Error("Cannot increment file counter", slog.String("user_id", userID), slog.String("file_id", fileID), slog.Any("error", err))

			return 0, fmt.Errorf("cannot increment file counter: %w", err)
		}

		return counter, nil
	}

	//FIXME: If the user exists, then the counter is incorrect.
	return 0, nil
}

func (d *downloadService) GetPage(ctx context.Context, id string) (string, error) {
	content, err := d.repo.GetPage(ctx, id)
	if err != nil {
		d.log.Error("Cannot get page content", slog.String("page_id", id), slog.Any("error", err))

		return "", fmt.Errorf("cannot get page %s content: %w", id, err)
	}

	return content, nil
}

func (d *downloadService) GetDownloadCounters(ctx context.Context, id string) (map[string]int, error) {
	counters, err := d.repo.GetDownloadCounters(ctx, id)
	if err != nil {
		d.log.Error("Cannot get download counters", slog.String("id", id), slog.Any("error", err))

		return nil, fmt.Errorf("cannot get download %s counters: %w", id, err)
	}

	return counters, nil
}
