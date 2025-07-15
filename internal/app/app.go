package app

import (
	"context"
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

type App struct {
	cfgPath string
	srv     *http.Server
}

func New(cfgPath string) *App {
	return &App{
		cfgPath: cfgPath,
	}
}

func (a *App) Start() {
	cfg := config.MustLoad(a.cfgPath)

	opt, err := redis.ParseURL(cfg.RedisURL)
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
	switch cfg.LogLevel {
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

	drepo, err := download.NewDownloadRepository(rdb, log)
	if err != nil {
		panic(err)
	}

	fsa, err := fsadapter.NewFSAdapter(
		cfg.IndexerConfig.WorkDir,
		cfg.IndexerConfig.DescFileName,
		cfg.IndexerConfig.TemplateFileName,
		cfg.HandlerConfig.URL, nil, log)
	if err != nil {
		panic(err)
	}

	store := index.NewIndexStorage(fsa, &cfg.IndexerConfig, log)
	iSrv := sindex.NewIndexService(store, drepo, log)
	dSrv := srvdownload.NewDownloadService(drepo, log)

	http.Handle("GET /share/{id}/{$}", httphandler.NewPageHandler(dSrv, log))
	http.Handle("GET /stat/{id}/{$}", httphandler.NewCounterHandler(dSrv, log))
	http.Handle("POST /file/{id}/{$}", httphandler.NewDownloadHandler(&cfg.HandlerConfig, dSrv, log))

	http.Handle("GET /index/{$}", httphandler.NewIndexHandler(iSrv, cfg.HandlerConfig.URL, log))

	a.srv = &http.Server{
		Addr: cfg.Listen,
	}

	go func() {
		log.Info("Start listen", slog.String("addr", cfg.Listen))

		if err := a.srv.ListenAndServe(); err != nil {
			log.Error("Could not serve", slog.String("listen_addr", cfg.Listen), slog.Any("error", err))
			os.Exit(2)
		}

	}()
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.srv.Shutdown(ctx)
}
