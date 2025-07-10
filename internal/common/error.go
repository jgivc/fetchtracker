package common

import "fmt"

var (
	ErrPageNotFoundError                = fmt.Errorf("page not found")
	ErrIndexingProcessHasAlreadyStarted = fmt.Errorf("indexing process has already started")
)
