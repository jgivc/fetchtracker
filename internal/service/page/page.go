package page

type PageRepository interface {
	GetPage(id string) (string, error)
}
type pageService struct {
}

func (p *pageService) GetPage(id string) (string, error) {
	panic("not implemented")
}
