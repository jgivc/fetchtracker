package httphandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/jgivc/fetchtracker/internal/common"
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
