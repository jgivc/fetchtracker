package entity

import "time"

// Download represents a single download (a folder). It's an aggregate.
type Download struct {
	ID          string // Stable hash, a unique identifier for the download
	Title       string // The title of the download from frontmatter, if available, or the folder name
	PageContent string // HTML description from description.md
	PageHash    string // ETag
	Enabled     bool
	Files       []*File   // The list of files belonging to this download
	SourcePath  string    // Internal path to the folder on the disk
	CreatedAt   time.Time // Creation time (of the first indexing)
}

type DownloadCounters struct {
	ID         string        `yaml:"id"`
	SourcePath string        `yaml:"path"`
	Files      []FileCounter `yaml:"files"`
}
