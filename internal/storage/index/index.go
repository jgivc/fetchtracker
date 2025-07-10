package index

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/jgivc/fetchtracker/internal/common"
	"github.com/jgivc/fetchtracker/internal/config"
	"github.com/jgivc/fetchtracker/internal/entity"
)

const (
	maxDirs = 100
)

type FSAdapter interface {
	ToDownload(folderPath string) (*entity.Download, error)
}

type indexStorage struct {
	running atomic.Bool
	adapter FSAdapter
	cfg     *config.IndexerConfig
	log     *slog.Logger
}

func NewIndexStorage(adapter FSAdapter, cfg *config.IndexerConfig, log *slog.Logger) *indexStorage {
	return &indexStorage{
		adapter: adapter,
		cfg:     cfg,
		log:     log.With(slog.String("item", "IndexStorage")),
	}
}

func (i *indexStorage) Scan(ctx context.Context) ([]*entity.Download, error) {
	if !i.running.CompareAndSwap(false, true) {
		return nil, common.ErrIndexingProcessHasAlreadyStarted
	}
	defer i.running.Store(false)

	entries, err := os.ReadDir(i.cfg.WorkDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(i.cfg.WorkDir, entry.Name()))
		}

		if len(dirs) >= maxDirs {
			break
		}
	}

	if len(dirs) == 0 {
		return []*entity.Download{}, nil
	}

	in := make(chan string, len(dirs))
	out := make(chan *entity.Download, len(dirs))

	for _, dir := range dirs {
		in <- dir
	}
	close(in)

	var wg sync.WaitGroup
	wg.Add(i.cfg.Workers)
	for n := 0; n < i.cfg.Workers; n++ {
		go i.worker(ctx, n, in, out, &wg)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	var downloads []*entity.Download
	for download := range out {
		i.log.Info("Found folder", slog.String("id", download.ID), slog.String("path", download.SourcePath))
		downloads = append(downloads, download)
	}

	return downloads, nil
}

func (i *indexStorage) worker(ctx context.Context, n int, in chan string, out chan *entity.Download, wg *sync.WaitGroup) {
	defer wg.Done()

	log := i.log.With(slog.Int("worker_id", n))
	log.Info("Started")

	for folderPath := range in {
		download, err := i.adapter.ToDownload(folderPath)
		if err != nil {
			log.Error("Cannot scan folder", slog.String("folder_path", folderPath), slog.Any("error", err))

			continue
		}

		select {
		case <-ctx.Done():
			log.Info("Interrupted")

			return
		case out <- download:
		}
	}

	log.Info("Done")
}
