package index

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/jgivc/fetchtracker/internal/common"
	"github.com/jgivc/fetchtracker/internal/entity"
)

type DownloadStorage interface {
	Scan(ctx context.Context) ([]*entity.Download, error)
}

type DownloadRepository interface {
	Save(ctx context.Context, downloads []*entity.Download) error
	Info(ctx context.Context) ([]*entity.ShareInfo, error)
	DownloadCounterIterator(ctx context.Context) (iter.Seq2[*entity.DownloadCounters, error], error)
}

type IndexerService struct {
	running atomic.Bool
	store   DownloadStorage
	repo    DownloadRepository
	log     *slog.Logger
}

func NewIndexService(store DownloadStorage, repo DownloadRepository, log *slog.Logger) *IndexerService {
	return &IndexerService{
		store: store,
		repo:  repo,
		log:   log.With(slog.String("item", "IndexService")),
	}
}

func (i *IndexerService) DumpCounters(ctx context.Context, path string) error {
	if !i.running.CompareAndSwap(false, true) {
		return common.ErrIndexingProcessHasAlreadyStarted
	}
	defer i.running.Store(false)

	i.log.Info("Dump counters")

	w, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create dump file: %w", err)
	}
	defer w.Close()

	it, err := i.repo.DownloadCounterIterator(ctx)
	if err != nil {
		return fmt.Errorf("cannot get iterator: %w", err)
	}

	encoder := json.NewEncoder(w)
	// encoder.SetIndent("", "  ")

	w.Write([]byte("[\n"))

	first := true
	for dc, err := range it {
		if err != nil {
			return fmt.Errorf("cannot get counters: %w", err)
		}

		if !first {
			w.Write([]byte(","))
		}
		first = false

		if err2 := encoder.Encode(dc); err2 != nil {
			return fmt.Errorf("cannot encode struct: %w", err2)
		}
	}

	w.Write([]byte("]"))

	return nil
}

func (i *IndexerService) Index(ctx context.Context) ([]*entity.ShareInfo, error) {
	if !i.running.CompareAndSwap(false, true) {
		return nil, common.ErrIndexingProcessHasAlreadyStarted
	}
	defer i.running.Store(false)

	i.log.Info("Start index process")

	downloads, err := i.store.Scan(ctx)
	if err != nil {
		i.log.Error("Cannot scan", slog.Any("error", err))

		return nil, fmt.Errorf("cannot scan download store: %w", err)
	}

	if len(downloads) < 1 {
		i.log.Error("Cannot find dirs")

		return nil, fmt.Errorf("cannot find dirs")
	}

	i.log.Info("Scan storage dirs", slog.Int("count", len(downloads)))

	if err := i.repo.Save(ctx, downloads); err != nil {
		i.log.Error("Cannot save scan content", slog.Any("error", err))

		return nil, fmt.Errorf("cannot save scan content: %w", err)
	}

	infos, err := i.repo.Info(ctx)
	if err != nil {
		i.log.Error("Cannot get file path", slog.Any("error", err))

		return nil, fmt.Errorf("cannot get download info: %w", err)
	}

	return infos, nil
}
