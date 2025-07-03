package index

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jgivc/fetchtracker/internal/entity"
)

const (
	serviceName = "index"
)

type DownloadStorage interface {
	Scan(ctx context.Context) ([]*entity.Download, error)
}

type DownloadRepository interface {
	Save(downloads []*entity.Download) error
}

type IndexerService struct {
	store DownloadStorage
	repo  DownloadRepository
	log   *slog.Logger
}

func NewIndexService(store DownloadStorage, repo DownloadRepository, log *slog.Logger) *IndexerService {
	return &IndexerService{
		store: store,
		repo:  repo,
		log:   log.With(slog.String("service", serviceName)),
	}
}

func (i *IndexerService) Index(ctx context.Context) error {
	downloads, err := i.store.Scan(ctx)
	if err != nil {
		i.log.Error("Cannot scan", slog.Any("error", err))

		return fmt.Errorf("cannot scan download store: %w", err)
	}

	if err := i.repo.Save(downloads); err != nil {
		i.log.Error("Cannot save scan content", slog.Any("error", err))

		return fmt.Errorf("cannot save scan content: %w", err)
	}

	return nil
}
