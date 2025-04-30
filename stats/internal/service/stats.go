package service

import (
	"context"
	"fmt"
	"log"
	"stats/internal/storage"
)

type StatsService struct {
	storage storage.Storage
}

func New(storage storage.Storage) *StatsService {
	return &StatsService{
		storage: storage}
}

func (s *StatsService) GetStats(ctx context.Context) (*storage.Stats, error) {
	const fn = "StatsService.GetStats"

	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	stats, err := tx.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	log.Print(stats.ID)

	return stats, nil
}
