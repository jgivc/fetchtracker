package entity

import "time"

// Download представляет одну раздачу (папку). Это агрегат.
type Download struct {
	ID          string // Стабильный хеш, уникальный идентификатор раздачи
	Title       string // Заголовок раздачи из frontmatter, если есть или имя папки
	PageContent string // HTML-описание из description.md
	PageHash    string // ETag
	Enabled     bool
	Files       []*File   // Список файлов, принадлежащих этой раздаче
	SourcePath  string    // Внутренний путь к папке на диске
	CreatedAt   time.Time // Время создания (первой индексации)
}

type DownloadCounters struct {
	ID         string        `yaml:"id"`
	SourcePath string        `yaml:"path"`
	Files      []FileCounter `yaml:"files"`
}
