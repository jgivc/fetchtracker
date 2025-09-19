package entity

// File represents a single downloadable file within a download.
type File struct {
	ID          string // A unique identifier for the file (e.g., a hash of the file path)
	Name        string // The name of the file
	Description string // A short description of the file, shown instead of Name if provided (from frontmatter).
	SourcePath  string // Internal path to the file on disk (relative to the download's root)
	URL         string
	Size        int64  // The size of the file in bytes
	MIMEType    string // The MIME type of the file
}

type FileCounter struct {
	ID         string `yaml:"id"`
	Name       string `yaml:"name"`
	SourcePath string `yaml:"path"`
	Counter    int64  `yaml:"counter"`
}
