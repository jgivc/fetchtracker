package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jgivc/fetchtracker/internal/adapter/fsadapter"
	"github.com/jgivc/fetchtracker/internal/config"
	httphandler "github.com/jgivc/fetchtracker/internal/handler/http"
	"github.com/jgivc/fetchtracker/internal/repository/download"
	srvdownload "github.com/jgivc/fetchtracker/internal/service/download"
	sindex "github.com/jgivc/fetchtracker/internal/service/index"
	"github.com/jgivc/fetchtracker/internal/storage/index"
	"github.com/redis/go-redis/v9"
)

const (
	indexTimeout = 5 * time.Second
	dumpTimeout  = 5 * time.Second
)

type App struct {
	cfgPath string
	cfg     *config.Config
	srv     *http.Server
	indexer *sindex.IndexerService
	log     *slog.Logger
}

func New(cfgPath string) *App {
	return &App{
		cfgPath: cfgPath,
	}
}

func (a *App) Start() {
	a.cfg = config.MustLoad(a.cfgPath)

	opt, err := redis.ParseURL(a.cfg.RedisURL)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(opt)
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	lo := &slog.HandlerOptions{}
	switch a.cfg.LogLevel {
	case config.LogLevelInfo:
		lo.Level = slog.LevelInfo
	case config.LogLevelWarn:
		lo.Level = slog.LevelWarn
	case config.LogLevelError:
		lo.Level = slog.LevelError
	case config.LogLevelDebug:
		lo.Level = slog.LevelDebug
	default:
		panic("unknown log level")
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, lo))
	a.log = log

	drepo, err := download.NewDownloadRepository(rdb, log)
	if err != nil {
		panic(err)
	}

	fsa, err := fsadapter.NewFSAdapter(a.cfg.FSAdapterConfig(), log)
	if err != nil {
		panic(err)
	}

	store := index.NewIndexStorage(fsa, &a.cfg.IndexerConfig, log)
	a.indexer = sindex.NewIndexService(store, drepo, log)
	dSrv := srvdownload.NewDownloadService(drepo, log)

	http.Handle("GET /share/{id}/{$}", httphandler.NewPageHandler(dSrv, log))
	http.Handle("GET /stat/{id}/{$}", httphandler.NewCounterHandler(dSrv, log))
	http.Handle("POST /file/{id}/{$}", httphandler.NewDownloadHandler(&a.cfg.HandlerConfig, dSrv, log))

	http.Handle("GET /index/{$}", httphandler.NewIndexHandler(a.indexer, a.cfg.HandlerConfig.URL, log))

	a.srv = &http.Server{
		Addr: a.cfg.Listen,
	}

	go func() {
		log.Info("Start listen", slog.String("addr", a.cfg.Listen))

		if err := a.srv.ListenAndServe(); err != nil {
			log.Error("Could not serve", slog.String("listen_addr", a.cfg.Listen), slog.Any("error", err))
			os.Exit(2)
		}

	}()
}

func (a *App) Dump() {
	ctx, cancel := context.WithTimeout(context.Background(), dumpTimeout)
	defer cancel()

	if err := a.indexer.DumpCounters(ctx, a.cfg.IndexerConfig.DumpFileName); err != nil {
		a.log.Error("Cannot dump counters", slog.Any("aeeoe", err))
	}
}

func (a *App) Index() {
	ctx, cancel := context.WithTimeout(context.Background(), indexTimeout)
	defer cancel()

	fmt.Println("Building...")

	infos, err := a.indexer.Index(ctx)
	if err != nil {
		fmt.Printf("Cannot build index: %s\n", err)

		return
	}

	for i, info := range infos {
		fmt.Printf("%d. %s -> %s/share/%s, files: %d\n", i+1, info.SourcePath, a.cfg.HandlerConfig.URL, info.ID, info.FileCount)
	}

	fmt.Println("Done.")
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.srv.Shutdown(ctx)
}
