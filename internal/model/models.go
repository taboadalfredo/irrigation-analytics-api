package model

import "time"

type Farm struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"not null"`
}

type IrrigationSector struct {
    ID     uint   `gorm:"primaryKey"`
    FarmID uint   `gorm:"not null;index"`
    Name   string `gorm:"not null"`
}

type IrrigationData struct {
    ID                 uint      `gorm:"primaryKey"`
    FarmID             uint      `gorm:"not null;index"`
    IrrigationSectorID uint      `gorm:"not null;index"`
    StartTime          time.Time `gorm:"not null;index"`
    EndTime            time.Time `gorm:"not null"`
    NominalAmount      float32
    RealAmount         float32
    Efficiency         float32
    CreatedAt          time.Time
    UpdatedAt          time.Time
}
