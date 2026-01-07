package service

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/yourorg/irrigation/internal/repository"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) DailyAggregation(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint) ([]repository.DailyRow, error) {
    args := m.Called(ctx, farmID, start, end, sectorID)
    return args.Get(0).([]repository.DailyRow), args.Error(1)
}

func (m *mockRepo) AggregatedMetrics(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint) (repository.AggregatedRow, error) {
    args := m.Called(ctx, farmID, start, end, sectorID)
    return args.Get(0).(repository.AggregatedRow), args.Error(1)
}

func TestServiceOK(t *testing.T) {
    repo := new(mockRepo)
    svc := NewAnalyticsService(repo)

    start := time.Now().Add(-24 * time.Hour)
    end := time.Now()

    repo.On("AggregatedMetrics", mock.Anything, uint(1), start, end, (*uint)(nil)).
        Return(repository.AggregatedRow{TotalNominalMM: 10, TotalEvents: 1}, nil)

    repo.On("DailyAggregation", mock.Anything, uint(1), start, end, (*uint)(nil)).
        Return([]repository.DailyRow{{EventCount: 1}}, nil)

    agg, ts, err := svc.GetDaily(context.Background(), 1, start, end, nil)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), agg.TotalEvents)
    assert.Len(t, ts, 1)
}
