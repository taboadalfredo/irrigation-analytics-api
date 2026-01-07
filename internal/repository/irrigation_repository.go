package repository

import (
    "context"
    "time"
    "gorm.io/gorm"
)

type TimeSeriesRow struct {
    Day time.Time
    NominalAmountMM float64
    RealAmountMM float64
    AverageEfficiency float64
    EventCount int64
}

type AggregatedRow struct {
    TotalNominalMM float64
    TotalEvents int64
    AverageEfficiency float64
    MinEfficiency float64
    MaxEfficiency float64
}

type IrrigationRepository interface {
    TimeSeriesAggregation(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint, aggregation Aggregation) ([]TimeSeriesRow, error)
    AggregatedMetrics(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint) (AggregatedRow, error)
}

type irrigationRepository struct {
    db *gorm.DB
}

func NewIrrigationRepository(db *gorm.DB) IrrigationRepository {
    return &irrigationRepository{db: db}
}

func (r *irrigationRepository) TimeSeriesAggregation(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint, aggregation Aggregation) ([]TimeSeriesRow, error) {
    var rows []TimeSeriesRow
    bucket := aggregationSQL(aggregation)

    q := r.db.WithContext(ctx).
        Table("irrigation_data").
        Select(bucket + " as day, SUM(nominal_amount) as nominal_amount_mm, SUM(real_amount) as real_amount_mm, AVG(efficiency) as average_efficiency, COUNT(*) as event_count").
        Where("farm_id = ? AND start_time >= ? AND start_time <= ?", farmID, start, end)

    if sectorID != nil {
        q = q.Where("irrigation_sector_id = ?", *sectorID)
    }

    err := q.Group("day").Order("day").Scan(&rows).Error
    return rows, err
}

func (r *irrigationRepository) AggregatedMetrics(ctx context.Context, farmID uint, start, end time.Time, sectorID *uint) (AggregatedRow, error) {
    var row AggregatedRow
    q := r.db.WithContext(ctx).
        Table("irrigation_data").
        Select("SUM(nominal_amount) as total_nominal_mm, COUNT(*) as total_events, AVG(efficiency) as average_efficiency, MIN(efficiency) as min_efficiency, MAX(efficiency) as max_efficiency").
        Where("farm_id = ? AND start_time >= ? AND start_time <= ?", farmID, start, end)

    if sectorID != nil {
        q = q.Where("irrigation_sector_id = ?", *sectorID)
    }

    err := q.Scan(&row).Error
    return row, err
}
