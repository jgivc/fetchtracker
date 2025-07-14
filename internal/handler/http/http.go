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
	"time"

	"github.com/google/uuid"
	"github.com/jgivc/fetchtracker/internal/common"
	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/jgivc/fetchtracker/internal/util"
)

const (
	downloadCookieName = "download_token"
	hdrUserAgent       = "User-Agent"

	prefixIDCookie      = "cookie"
	prefixIDFingerpring = "fingerprint" // User-Agent + ip
)

var (
	idRegexp     = regexp.MustCompile(`^[a-f\d]{40}$`)
	cookieRegexp = regexp.MustCompile(`^[a-f\d\-]{36}$`)
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

type DownloadService interface {
	Download(ctx context.Context, id string) (string, error)
	IncFileCounter(ctx context.Context, userID, fileID string) (int64, error)
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

	getUserID := func(r *http.Request) string {
		cookie, err := r.Cookie(downloadCookieName)
		if err == nil {
			if cookieRegexp.MatchString(cookie.Value) {
				log.Info("Cookie found", slog.String("cookie", cookie.Value))
				return cookie.Value
			}
		}

		if err != nil && err != http.ErrNoCookie {
			log.Error("Cannot get user cookie", slog.Any("error", err))

		}

		uid := uuid.New().String()
		log.Info("Set new cookie", slog.String("cookie", uid))

		return uid
	}

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

		uid := getUserID(r)
		cookie := http.Cookie{
			Name:     downloadCookieName,
			Path:     "/",
			Value:    uid,
			Expires:  time.Now().Add(24 * time.Hour * 365), // Cookie expires in 1 year
			HttpOnly: true,                                 // Prevents JavaScript access (XSS protection)
			Secure:   true,                                 // Only send over HTTPS
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)

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

func NewDownloadHandler(hdrRedirect string, hdrRealIP string, srv DownloadService, log *slog.Logger) http.HandlerFunc {
	log = log.With(slog.String("handler", "DownloadHandler"))

	getUserID := func(r *http.Request) string {
		cookie, err := r.Cookie(downloadCookieName)
		if err == nil {
			if cookieRegexp.MatchString(cookie.Value) {
				log.Info("Cookie found", slog.String("cookie", cookie.Value))
				return fmt.Sprintf("%s:%s", prefixIDCookie, cookie.Value)
			}
		}

		if err != nil {
			fmt.Println("GGG", err)
		}

		if err != nil && err != http.ErrNoCookie {
			log.Error("Cannot get user cookie", slog.Any("error", err))

		}

		fp := fmt.Sprintf("%s:%s", r.Header.Get(hdrRealIP), r.Header.Get(hdrUserAgent))
		uid := util.GetIDFromString(&fp)

		log.Info("Cannot find cookie", slog.String("fingerprint", uid))

		return fmt.Sprintf("%s:%s", prefixIDFingerpring, uid)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		fileID := r.PathValue("id")
		if !idRegexp.MatchString(fileID) {
			http.Error(w, "Bad request", http.StatusBadRequest)

			return
		}

		log = log.With("remote_addr", r.Header.Get(hdrRealIP), slog.String("file_id", fileID))
		log.Info("New download request")

		//FIXME: For errors you need to answer something to the user
		path, err := srv.Download(context.Background(), fileID)
		if err != nil {
			switch {
			case errors.Is(err, common.ErrFileNotFoundError):
				http.Error(w, "Cannot find file", http.StatusNotFound)
			default:
				http.Error(w, "Cannot get file", http.StatusInternalServerError)
			}

			return
		}

		uid := getUserID(r)
		fmt.Println("AAA", uid)

		counter, err := srv.IncFileCounter(context.Background(), getUserID(r), fileID)
		if err != nil {
			http.Error(w, "Cannot get file", http.StatusInternalServerError)

			return
		}

		log.Info("Download file", slog.String("id", fileID), slog.String("path", path), slog.Int64("counter", counter))

		w.Header().Set(hdrRedirect, path)
	}
}

func getUserID(r *http.Request) string {
	panic("not implemented")
}
