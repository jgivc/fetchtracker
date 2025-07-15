package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jgivc/fetchtracker/internal/adapter/fsadapter"
	"github.com/jgivc/fetchtracker/internal/config"
	httphandler "github.com/jgivc/fetchtracker/internal/handler/http"
	"github.com/jgivc/fetchtracker/internal/repository/download"
	srvdownload "github.com/jgivc/fetchtracker/internal/service/download"
	sindex "github.com/jgivc/fetchtracker/internal/service/index"
	"github.com/jgivc/fetchtracker/internal/storage/index"
	"github.com/joho/godotenv"
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
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	db := os.Getenv("REDIS_DB")
	dbNum, _ := strconv.Atoi(db)
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       dbNum,                       // use default DB
	})

	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	drepo, err := download.NewDownloadRepository(rdb, log)
	if err != nil {
		panic(err)
	}

	cfg := &config.Config{
		URL:            "http://127.0.0.1:10011",
		Listen:         ":10011",
		RedirectHeader: config.RedirectHeader,
		RealIPHeader:   config.RealIPHeader,
		IndexerConfig: config.IndexerConfig{
			WorkDir:      "/tmp/testdata/",
			Workers:      2,
			DescFileName: "description.yml",
		},

		// TemplateFileName: "/tmp/template.txt",
	}

	fsa, err := fsadapter.NewFSAdapter(cfg.IndexerConfig.WorkDir, cfg.IndexerConfig.DescFileName, cfg.IndexerConfig.TemplateFileName, cfg.URL, nil, log)
	if err != nil {
		panic(err)
	}
	store := index.NewIndexStorage(fsa, &cfg.IndexerConfig, log)

	iSrv := sindex.NewIndexService(store, drepo, log)

	// iSrv.Index(ctx)

	// pSrv := page.NewPageService(drepo, log)
	// cSrv := counter.NewCounterService(drepo, log)
	dSrv := srvdownload.NewDownloadService(drepo, log)

	http.Handle("GET /share/{id}/{$}", httphandler.NewPageHandler(dSrv, log))
	http.Handle("GET /stat/{id}/{$}", httphandler.NewCounterHandler(dSrv, log))
	http.Handle("GET /info/{$}", httphandler.NewInfoHandler(cfg.URL, dSrv, log))
	http.Handle("POST /file/{id}/{$}", httphandler.NewDownloadHandler(cfg.RedirectHeader, cfg.RealIPHeader, dSrv, log))

	http.Handle("GET /index/{$}", httphandler.NewIndexHandler(iSrv, log))

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
