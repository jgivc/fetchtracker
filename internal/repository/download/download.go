package download

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"sync/atomic"

	"github.com/jgivc/fetchtracker/internal/common"
	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/redis/go-redis/v9"
)

const (
	KeyVersion1         = "v1"
	KeyVersion2         = "v2"
	KeyActiveVersion    = "av"  // STRING.
	KeyDownloadMap      = "dm"  // HASH. download_map:ver folder_id: folder_path
	KeyFilesMap         = "fm"  // HASH. files_map:ver file_id: file_path
	KeyDownloadFilesMap = "dfm" // HASH. download_files_map:ver:folder_id file_id: file_path
	// KeyDownloadMap   = "download_map"   // HASH. Maps the stable hash of a distribution to its path in the file system. HGET download_map:v1 {хеш_раздачи} -> /path/to/folder
	KeyPageContent = "pc" // HASH. {хеш_раздачи} -> HTML
	// KeyDownloadVersion = "download_versions" // HASH. Maps the stable hash of a distribution to the hash of its page content (ETag). HGET download_versions:v1 {distribution_hash} -> {content_hash}
	// KeyPageContent = "page_content" // STRING. Stores the full, ready-to-be-distributed HTML code of the distribution page. The key is an ETag.

	KeyFileStats      = "fs" // HASH. Key storage of statistics. Maps a stable hash of a file to its counter. Allows atomic increment. HINCRBY file_stats {file_hash} 1
	KeyUniqueDownload = "dl" // STRING. Used to cut off duplicate downloads. The key is the user ID (cookie/fingerprint). Set via SETNX with EX (TTL).

	KeyEmpty     = ""
	KeySeparator = ":"

	ScanCount                 = 1000
	defaultDownloadExpiration = 24 * time.Hour
)

var (
	// ClearableKeys = []string{KeyDownloadMap, KeyDownloadVersion, KeyPageContent}
	ClearableKeys = []string{KeyDownloadMap, KeyFilesMap, KeyDownloadFilesMap, KeyPageContent}
)

type downloadRepository struct {
	ver atomic.Value
	cl  *redis.Client
	log *slog.Logger
}

func NewDownloadRepository(cl *redis.Client, log *slog.Logger) (*downloadRepository, error) {

	repo := &downloadRepository{
		cl:  cl,
		log: log.With(slog.String("item", "DownloadRepository")),
	}

	ver, _, err := repo.getVersions(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot get active sersionL %w", err)
	}

	repo.ver.Store(ver)

	return repo, nil
}

func (r *downloadRepository) Info(ctx context.Context) ([]*entity.ShareInfo, error) {
	ver := r.getActiveVersion()

	downloadMap, err := r.cl.HGetAll(ctx, getKey(KeyDownloadMap, ver)).Result()
	if err != nil {
		return nil, fmt.Errorf("cannot get download map: %w", err)
	}

	if len(downloadMap) < 1 {
		return nil, common.ErrNoDownloadsFoundError
	}

	infos := make([]*entity.ShareInfo, 0, len(downloadMap))
	for id, path := range downloadMap {
		files, err := r.cl.HGetAll(ctx, getKey(KeyDownloadFilesMap, ver, id)).Result()
		if err != nil {
			return nil, fmt.Errorf("cannot get download files: %w", err)
		}

		infos = append(infos, &entity.ShareInfo{
			ID:         id,
			SourcePath: path,
			FileCount:  len(files),
		})
	}

	return infos, nil
}

func (r *downloadRepository) Save(ctx context.Context, downloads []*entity.Download) error {
	verActive, verStandby, err := r.getVersions(ctx)
	if err != nil {
		r.log.Error("Cannot get standby data version")

		return fmt.Errorf("cannot get active version: %w", err)
	}
	r.log.Info("Save new data", slog.String("active_version", verActive), slog.String("standby_version", verStandby))

	if err := r.clearOldData(ctx, verStandby); err != nil {
		r.log.Error("Cannot clear old data", slog.String("version", verStandby), slog.Any("error", err))

		return fmt.Errorf("cannot clear old data: %w", err)
	}

	if err := r.saveNewData(ctx, verStandby, downloads); err != nil {
		r.log.Error("Cannot save new data", slog.String("version", verStandby), slog.Any("error", err))

		return fmt.Errorf("cannot save new data: %w", err)
	}

	_, err = r.cl.Set(ctx, KeyActiveVersion, verStandby, 0).Result()
	if err != nil {
		r.log.Error("Cannot switch to new version", slog.String("version", verStandby), slog.Any("error", err))

		return fmt.Errorf("cannot switch to new version: %w", err)
	}

	r.ver.Store(verStandby)

	if err := r.clearDeletedFileCounters(ctx, downloads); err != nil {
		r.log.Error("Cannot delete deleted keys", slog.String("version", verStandby), slog.Any("error", err))

		return fmt.Errorf("cannot delete deleted keys: %w", err)
	}

	return nil
}

func (r *downloadRepository) clearDeletedFileCounters(ctx context.Context, downloads []*entity.Download) error {
	filesMap := make(map[string]struct{})
	for _, download := range downloads {
		for _, file := range download.Files {
			filesMap[file.ID] = struct{}{}
		}
	}

	pattern := getKey(KeyFileStats, "*")
	var (
		cursor       uint64
		deletedCount int64
	)

	pipe := r.cl.Pipeline()

	for {
		keys, nextCursor, err := r.cl.Scan(ctx, cursor, pattern, ScanCount).Result()
		if err != nil {
			return fmt.Errorf("error scanning keys: %w", err)
		}

		if len(keys) > 0 {
			for _, key := range keys {
				if _, exists := filesMap[key]; !exists {
					pipe.Del(ctx, key)
					deletedCount++
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if deletedCount > 0 {
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("cannot delete keys: %w", err)
		}
	}

	return nil
}

func (r *downloadRepository) saveNewData(ctx context.Context, ver string, downloads []*entity.Download) error {
	log := r.log.With(slog.String("op", "saveNewData"), slog.String("version", ver))
	log.Info("Save new data")

	pipe := r.cl.Pipeline()
	for _, download := range downloads {
		// pipe.HSet(ctx, getKey(KeyDownloadMap, ver), download.ID, download.SourcePath)
		pipe.HSet(ctx, getKey(KeyDownloadMap, ver), download.ID, download.SourcePath)
		pipe.HSet(ctx, getKey(KeyPageContent, ver), download.ID, download.PageContent)
		keyFileMap := getKey(KeyFilesMap, ver)
		keyDownloadMap := getKey(KeyDownloadFilesMap, ver, download.ID)
		for _, file := range download.Files {
			// pipe.HSet(ctx, keyFileMap, file.ID, file.SourcePath)
			// pipe.HSet(ctx, keyDownloadMap, file.ID, file.SourcePath)
			pipe.HSet(ctx, keyFileMap, file.ID, file.URL)
			pipe.HSet(ctx, keyDownloadMap, file.ID, file.URL)
		}
		// pipe.HSet(ctx, getKey(KeyDownloadVersion, ver), download.ID, download.PageHash)
		// pipe.Set(ctx, getKey(KeyPageContent, ver, download.PageHash), download.PageContent, 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("cannot save new data: %w", err)
	}

	return nil
}

func (r *downloadRepository) clearOldData(ctx context.Context, ver string) error {
	log := r.log.With(slog.String("op", "clearOldData"), slog.String("version", ver))
	log.Info("Clear old data")

	for _, key := range ClearableKeys {
		pattern := getKey(key, ver, "*")

		log.Info("Clear keys", slog.String("pattern", pattern))

		var (
			cursor       uint64
			deletedCount int64
		)

		for {

			keys, nextCursor, err := r.cl.Scan(ctx, cursor, pattern, ScanCount).Result()
			if err != nil {
				return fmt.Errorf("error scanning keys: %w", err)
			}

			if len(keys) > 0 {
				delCmd := r.cl.Del(ctx, keys...)
				count, err := delCmd.Result()
				if err != nil {
					return fmt.Errorf("error deleting keys: %w", err)
				}
				deletedCount += count
			}

			cursor = nextCursor
			if cursor == 0 {
				break // No more keys to scan
			}
		}

		_, err := r.cl.Del(ctx, getKey(key, ver)).Result()
		if err != nil {
			return fmt.Errorf("error deleting keys: %w", err)
		}

		log.Info("Clear keys", slog.String("pattern", pattern), slog.Int64("key_count", deletedCount))
	}

	return nil
}

/*
getVersions return active and standby versions
*/
func (r *downloadRepository) getVersions(ctx context.Context) (string, string, error) {
	ver, err := r.cl.Get(ctx, KeyActiveVersion).Result()
	if err != nil && err != redis.Nil {
		return KeyEmpty, KeyEmpty, fmt.Errorf("cannot get active version: %w", err)
	}

	switch ver {
	case KeyVersion1:
		return KeyVersion1, KeyVersion2, nil
	case KeyVersion2:
		return KeyVersion2, KeyVersion1, nil
	}

	r.log.Info("Active version key is not found. Try to set new one", slog.String("version", KeyVersion1))

	if _, err = r.cl.Set(ctx, KeyActiveVersion, KeyVersion1, 0).Result(); err != nil {
		return KeyEmpty, KeyEmpty, fmt.Errorf("cannot set varsion key: %w", err)
	}

	return KeyVersion1, KeyVersion2, nil
}

func (r *downloadRepository) GetFilePath(ctx context.Context, id string) (string, error) {
	path, err := r.cl.HGet(ctx, getKey(KeyFilesMap, r.getActiveVersion()), id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrFileNotFoundError
		}

		return "", fmt.Errorf("cannot get file %s path: %w", id, err)
	}

	return path, nil
}

func (r *downloadRepository) UserExists(ctx context.Context, id string) (bool, error) {
	res, err := r.cl.SetNX(ctx, getKey(KeyUniqueDownload, id), "1", defaultDownloadExpiration).Result()
	if err != nil {
		return false, fmt.Errorf("cannot check user exists")
	}

	return !res, nil
}

func (r *downloadRepository) IncFileCounter(ctx context.Context, id string) (int64, error) {
	counter, err := r.cl.HIncrBy(ctx, KeyFileStats, id, 1).Result()
	if err != nil {
		return 0, fmt.Errorf("cannot increment file %s counter: %w", id, err)
	}

	return counter, nil
}

func (r *downloadRepository) GetDownloadCounters(ctx context.Context, id string) (map[string]int, error) {
	files, err := r.cl.HGetAll(ctx, getKey(KeyDownloadFilesMap, r.getActiveVersion(), id)).Result()
	if err != nil {
		return nil, fmt.Errorf("cannot get download files: %w", err)
	}

	counters := make(map[string]int)
	for fileID := range files {
		counter, err := r.cl.HGet(ctx, KeyFileStats, fileID).Result()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				r.log.Error("cannot get counter for file", slog.String("file_id", fileID), slog.Any("error", err))
			} else {
				counters[fileID] = 0
			}

			continue
		}

		c, err := strconv.Atoi(counter)
		if err != nil {
			r.log.Error("cannot convert counter to int", slog.String("file_id", fileID), slog.Any("error", err))

			continue
		}

		counters[fileID] = c
	}

	return counters, nil
}

func (r *downloadRepository) getActiveVersion() string {
	return r.ver.Load().(string)
}

func (r *downloadRepository) GetPage(ctx context.Context, id string) (string, error) {
	// str, err := r.cl.Get(ctx, getKey(KeyPageContent, r.getActiveVersion())).Result()
	str, err := r.cl.HGet(ctx, getKey(KeyPageContent, r.getActiveVersion()), id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrPageNotFoundError
		}

		return "", err
	}

	return str, nil
}

func (r *downloadRepository) DownloadCounterIterator(ctx context.Context) (iter.Seq2[*entity.DownloadCounters, error], error) {
	ver := r.getActiveVersion()
	folders, err := r.cl.HGetAll(ctx, getKey(KeyDownloadMap, ver)).Result()
	if err != nil {
		return nil, fmt.Errorf("cannot getfolder list: %w", err)
	}

	return func(yield func(*entity.DownloadCounters, error) bool) {
		for folderID, folderPath := range folders {
			dc := &entity.DownloadCounters{
				ID:         folderID,
				SourcePath: folderPath,
			}

			filesMap, err := r.cl.HGetAll(ctx, getKey(KeyDownloadFilesMap, ver, folderID)).Result()
			if err != nil {
				yield(nil, fmt.Errorf("cannot get folder files: %w", err))

				return
			}

			if len(filesMap) < 1 {
				if !yield(dc, nil) {
					return
				}

				continue
			}

			fileCounters := make([]entity.FileCounter, 0, len(filesMap))

			pipe := r.cl.Pipeline()
			for fileID, filePath := range filesMap {
				fileCounters = append(fileCounters, entity.FileCounter{
					ID:         fileID,
					SourcePath: filePath,
				})
				pipe.HGet(ctx, KeyFileStats, fileID)
			}

			cmds, err := pipe.Exec(ctx)
			if err != nil {
				yield(nil, fmt.Errorf("cannot exec pipe: %w", err))

				return
			}

			for i, cmd := range cmds {
				var counter int64
				val, err := cmd.(*redis.StringCmd).Result()
				if err != nil {
					if err != redis.Nil {
						r.log.Error("cannot get file counter", slog.Any("error", err))
					}
				} else {
					counter, err = strconv.ParseInt(val, 10, 64)
					if err != nil {
						r.log.Error("Cannot convert counter value to string", slog.Any("error", err))
						counter = 0
					}
				}

				fileCounters[i].Counter = counter

			}

			dc.Files = fileCounters

			if !yield(dc, nil) {
				return
			}
		}
	}, nil
}

func getKey(keys ...string) string {
	return strings.Join(keys, KeySeparator)
}
