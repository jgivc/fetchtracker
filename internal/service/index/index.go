package index

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jgivc/fetchtracker/internal/entity"
)

type DownloadStorage interface {
	Scan(ctx context.Context) ([]*entity.Download, error)
}

type DownloadRepository interface {
	Save(ctx context.Context, downloads []*entity.Download) error
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
		log:   log.With(slog.String("item", "IndexService")),
	}
}

func (i *IndexerService) Index(ctx context.Context) error {
	downloads, err := i.store.Scan(ctx)
	if err != nil {
		i.log.Error("Cannot scan", slog.Any("error", err))

		return fmt.Errorf("cannot scan download store: %w", err)
	}

	if len(downloads) < 1 {
		i.log.Error("Cannot find dirs")

		return fmt.Errorf("cannot find dirs")
	}

	i.log.Info("Scan storage dirs", slog.Int("count", len(downloads)))

	if err := i.repo.Save(ctx, downloads); err != nil {
		i.log.Error("Cannot save scan content", slog.Any("error", err))

		return fmt.Errorf("cannot save scan content: %w", err)
	}

	return nil
}
