package repository

type Aggregation string

const (
    AggregationDaily   Aggregation = "daily"
    AggregationWeekly  Aggregation = "weekly"
    AggregationMonthly Aggregation = "monthly"
)

func IsValidAggregation(a Aggregation) bool {
    switch a {
    case AggregationDaily, AggregationWeekly, AggregationMonthly:
        return true
    default:
        return false
    }
}

func aggregationSQL(a Aggregation) string {
    switch a {
    case AggregationWeekly:
        return "date_trunc('week', start_time)"
    case AggregationMonthly:
        return "date_trunc('month', start_time)"
    default:
        return "date_trunc('day', start_time)"
    }
}
