package index

import (
	"context"
	"fmt"

	"github.com/jgivc/fetchtracker/internal/entity"
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
}

func (i *IndexerService) Index(ctx context.Context) error {
	downloads, err := i.store.Scan(ctx)
	if err != nil {
		return fmt.Errorf("cannot scan download store: %w", err)
	}

	return i.repo.Save(downloads)
}
