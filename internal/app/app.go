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
	"github.com/jgivc/fetchtracker/internal/service/counter"
	sindex "github.com/jgivc/fetchtracker/internal/service/index"
	"github.com/jgivc/fetchtracker/internal/service/page"
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

	cfg := &config.IndexerConfig{
		URL:          "127.0.0.1",
		Listen:       ":10011",
		WorkDir:      "testdata/",
		Workers:      2,
		DescFileName: "description.yml",
		// TemplateFileName: "/tmp/template.txt",
	}

	fsa, err := fsadapter.NewFSAdapter(cfg.DescFileName, cfg.TemplateFileName, cfg.URL, nil, log)
	if err != nil {
		panic(err)
	}
	store := index.NewIndexStorage(fsa, cfg, log)

	iSrv := sindex.NewIndexService(store, drepo, log)

	// iSrv.Index(ctx)

	pSrv := page.NewPageService(drepo, log)
	cSrv := counter.NewCounterService(drepo, log)

	http.Handle("GET /share/{id}/", httphandler.NewPageHandler(pSrv, log))
	http.Handle("GET /index/{$}", httphandler.NewIndexHandler(iSrv, log))
	http.Handle("GET /info/{id}/", httphandler.NewCounterHandler(cSrv, log))

	a.srv = &http.Server{
		Addr: cfg.Listen,
	}

	go func() {
		log.Info("Start lisnen", slog.String("addr", cfg.Listen))

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
