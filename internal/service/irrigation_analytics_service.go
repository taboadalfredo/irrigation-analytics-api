package service

import (
    "context"
    "time"
    "github.com/yourorg/irrigation/internal/repository"
)

type AnalyticsService struct {
    repo repository.IrrigationRepository
}

func NewAnalyticsService(r repository.IrrigationRepository) *AnalyticsService {
    return &AnalyticsService{repo: r}
}

func (s *AnalyticsService) GetDaily(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint) (repository.AggregatedRow, []repository.DailyRow, error) {
    agg, err := s.repo.AggregatedMetrics(ctx, farmID, start, end, sectorID)
    if err != nil {
        return repository.AggregatedRow{}, nil, err
    }
    ts, err := s.repo.DailyAggregation(ctx, farmID, start, end, sectorID)
    return agg, ts, err
}
