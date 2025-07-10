package page

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	serviceName = "page"
)

type PageRepository interface {
	GetPage(ctx context.Context, id string) (string, error)
}

type pageService struct {
	repo PageRepository
	log  *slog.Logger
}

func NewPageService(repo PageRepository, log *slog.Logger) *pageService {
	return &pageService{
		repo: repo,
		log:  log.With(slog.String("service", serviceName)),
	}
}

func (p *pageService) GetPage(ctx context.Context, id string) (string, error) {
	content, err := p.repo.GetPage(ctx, id)
	if err != nil {
		p.log.Error("Cannot get page content", slog.String("page_id", id), slog.Any("error", err))

		return "", fmt.Errorf("cannot get page %s content: %w", id, err)
	}

	return content, nil
}
