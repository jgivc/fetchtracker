package common

import "fmt"

var (
	ErrPageNotFoundError                = fmt.Errorf("page not found")
	ErrFileNotFoundError                = fmt.Errorf("file not found")
	ErrIndexingProcessHasAlreadyStarted = fmt.Errorf("indexing process has already started")
	ErrNoDownloadsFoundError            = fmt.Errorf("no downloads found")
)
