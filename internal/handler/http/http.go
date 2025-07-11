package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/jgivc/fetchtracker/internal/common"
	"github.com/jgivc/fetchtracker/internal/entity"
)

var (
	idRegexp = regexp.MustCompile(`^[a-f\d]{40}$`)
)

type PageService interface {
	GetPage(ctx context.Context, id string) (string, error)
}

type IndexService interface {
	Index(ctx context.Context) error
}

type CounterService interface {
	GetDownloadCounters(ctx context.Context, id string) (map[string]int, error)
}

type InfoService interface {
	Info(ctx context.Context) ([]*entity.ShareInfo, error)
}

func NewIndexHandler(srv IndexService, log *slog.Logger) http.HandlerFunc {
	log = log.With(slog.String("handler", "IndexHandler"))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := srv.Index(context.Background()); err != nil {
			switch {
			case errors.Is(err, common.ErrIndexingProcessHasAlreadyStarted):
				http.Error(w, "Index process has already started", http.StatusConflict)
			default:
				http.Error(w, "Cannot start index process", http.StatusInternalServerError)
			}

			return
		}

		w.Write([]byte("done"))
	}
}

func NewPageHandler(srv PageService, log *slog.Logger) http.HandlerFunc {
	log = log.With(slog.String("handler", "PageHandler"))

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !idRegexp.MatchString(id) {
			http.Error(w, "Bad request", http.StatusBadRequest)

			return
		}

		content, err := srv.GetPage(context.Background(), id)
		if err != nil {
			switch {
			case errors.Is(err, common.ErrPageNotFoundError):
				http.Error(w, "Cannot get page", http.StatusNotFound)
			default:
				http.Error(w, "Cannot get page", http.StatusInternalServerError)
			}

			return
		}

		w.Write([]byte(content))
	}
}

func NewCounterHandler(srv CounterService, log *slog.Logger) http.HandlerFunc {
	log = log.With(slog.String("handler", "CounterHandler"))

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !idRegexp.MatchString(id) {
			http.Error(w, "Bad request", http.StatusBadRequest)

			return
		}

		counters, err := srv.GetDownloadCounters(context.Background(), id)
		if err != nil {
			http.Error(w, "Cannot get page", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(counters); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func NewInfoHandler(siteURL string, srv InfoService, log *slog.Logger) http.HandlerFunc {
	log = log.With(slog.String("handler", "InfoHandler"))

	return func(w http.ResponseWriter, r *http.Request) {
		infos, err := srv.Info(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		buf := bytes.Buffer{}

		for _, info := range infos {
			buf.WriteString(fmt.Sprintf("%s -> %s/share/%s, files: %d\n", info.SourcePath, siteURL, info.ID, info.FileCount))
		}

		w.Write(buf.Bytes())
	}
}
