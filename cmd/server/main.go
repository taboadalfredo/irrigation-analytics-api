package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/yourorg/irrigation/internal/controllers"
    "github.com/yourorg/irrigation/internal/model"
    "github.com/yourorg/irrigation/internal/repository"
    "github.com/yourorg/irrigation/internal/service"
)

func main() {
    dsn := "host=" + os.Getenv("DB_HOST") +
        " user=" + os.Getenv("DB_USER") +
        " password=" + os.Getenv("DB_PASSWORD") +
        " dbname=" + os.Getenv("DB_NAME") +
        " port=" + os.Getenv("DB_PORT") +
        " sslmode=disable"

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    if err := db.AutoMigrate(
        &model.Farm{},
        &model.IrrigationSector{},
        &model.IrrigationData{},
    ); err != nil {
        log.Fatal(err)
    }

    repo := repository.NewIrrigationRepository(db)
    svc := service.NewAnalyticsService(repo)
    ctrl := controllers.NewIrrigationAnalyticsController(svc)

    r := gin.Default()
    r.GET("/v1/farms/:farm_id/irrigation/analytics", ctrl.GetAnalytics)

    log.Println("listening on :8080")
    log.Fatal(r.Run(":8080"))
}
