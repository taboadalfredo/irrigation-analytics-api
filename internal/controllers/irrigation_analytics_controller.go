package controllers

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/yourorg/irrigation/internal/repository"
    "github.com/yourorg/irrigation/internal/service"
)

type IrrigationAnalyticsController struct {
    svc *service.AnalyticsService
}

func NewIrrigationAnalyticsController(svc *service.AnalyticsService) *IrrigationAnalyticsController {
    return &IrrigationAnalyticsController{svc: svc}
}

func (c *IrrigationAnalyticsController) GetAnalytics(ctx *gin.Context) {
    farmID64, err := strconv.ParseUint(ctx.Param("farm_id"), 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid farm_id"})
        return
    }
    farmID := uint(farmID64)

    end := time.Now().UTC()
    start := end.Add(-7 * 24 * time.Hour)

    if v := ctx.Query("start_date"); v != "" {
        if t, err := time.Parse(time.RFC3339, v); err == nil {
            start = t
        }
    }
    if v := ctx.Query("end_date"); v != "" {
        if t, err := time.Parse(time.RFC3339, v); err == nil {
            end = t
        }
    }

    var sectorID *uint
    if v := ctx.Query("sector_id"); v != "" {
        if id, err := strconv.ParseUint(v, 10, 64); err == nil {
            u := uint(id)
            sectorID = &u
        }
    }

    aggregation := repository.Aggregation(ctx.DefaultQuery("aggregation", "daily"))
    if !repository.IsValidAggregation(aggregation) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid aggregation"})
        return
    }

    resp, err := c.svc.GetAnalyticsWithYoY(ctx, farmID, start, end, sectorID, aggregation)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, resp)
}
