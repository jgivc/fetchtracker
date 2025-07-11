package download

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jgivc/fetchtracker/internal/entity"
)

const (
	serviceName = "download"
)

type DownloadRepository interface {
	GetFilePath(ctx context.Context, id string) (string, error)
	IncFileCounter(ctx context.Context, id string) (int64, error)
	Info(ctx context.Context) ([]*entity.ShareInfo, error)
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

func (d *downloadService) IncFileCounter(ctx context.Context, id string) (int64, error) {
	counter, err := d.repo.IncFileCounter(ctx, id)
	if err != nil {
		d.log.Error("Cannot increment file counter", slog.String("file_id", id), slog.Any("error", err))

		return 0, fmt.Errorf("cannot increment file counter: %w", err)
	}

	return counter, nil
}

func (d *downloadService) Info(ctx context.Context) ([]*entity.ShareInfo, error) {
	infos, err := d.repo.Info(ctx)
	if err != nil {
		d.log.Error("Cannot get file path", slog.Any("error", err))

		return nil, fmt.Errorf("cannot get download info: %w", err)
	}

	return infos, nil
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
