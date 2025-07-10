package counter

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	serviceName = "counter"
)

type CounterRepository interface {
	GetDownloadCounters(ctx context.Context, id string) (map[string]int, error)
}

type counterService struct {
	repo CounterRepository
	log  *slog.Logger
}

func NewCounterService(repo CounterRepository, log *slog.Logger) *counterService {
	return &counterService{
		repo: repo,
		log:  log.With(slog.String("service", serviceName)),
	}
}

func (c *counterService) GetDownloadCounters(ctx context.Context, id string) (map[string]int, error) {
	counters, err := c.repo.GetDownloadCounters(ctx, id)
	if err != nil {
		c.log.Error("Cannot get download counters", slog.String("id", id), slog.Any("error", err))

		return nil, fmt.Errorf("cannot get download %s counters: %w", id, err)
	}

	return counters, nil
}
