package entity

// File представляет один скачиваемый файл внутри раздачи.
type File struct {
	ID          string // Уникальный идентификатор файла (например, хеш пути файла)
	Name        string // "Красивое" имя для отображения (из manifest.yml)
	Description string // Краткое описание файла, выводится вместо Name, если указано(frontmatted).
	SourcePath  string // Внутренний путь к файлу на диске (относительно корня раздачи)
	URL         string
	Size        int64  // Размер файла в байтах
	MIMEType    string // MIME-тип файла
}

type FileCounter struct {
	ID         string `yaml:"id"`
	Name       string `yaml:"name"`
	SourcePath string `yaml:"path"`
	Counter    int64  `yaml:"counter"`
}
