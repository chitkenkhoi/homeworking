package models

import (
    "time"

    "gorm.io/gorm"
)

type Sprint struct {
    ID        int           `gorm:"primaryKey;autoIncrement" json:"id"`
    CreatedAt time.Time     `json:"created_at"`
    UpdatedAt time.Time     `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Name       string    `gorm:"not null;size:255" json:"name"`
    StartDate  time.Time `gorm:"not null" json:"start_date"`
    EndDate    time.Time `gorm:"not null" json:"end_date"`
    ProjectID  int       `gorm:"index;not null" json:"project_id"`
    Goal       string    `gorm:"type:text" json:"goal"`

    Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"project"`
    Tasks []Task `gorm:"foreignKey:SprintID" json:"tasks,omitempty"`
}