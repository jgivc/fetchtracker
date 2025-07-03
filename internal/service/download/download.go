package download

import "github.com/jgivc/fetchtracker/internal/entity"

type DownloadRepository interface {
	Save(downloads []*entity.Download) error
}

type downloadService struct {
}

func (d *downloadService) Download(id string) (string, error) {
	panic("not implemented")
}
