package models

import (
	"time"

	"gorm.io/gorm"
)

type ProjectStatus string

const (
	StatusActive    ProjectStatus = "ACTIVE"
	StatusCompleted ProjectStatus = "COMPLETED"
	StatusOnHold    ProjectStatus = "ON_HOLD"
	StatusCancelled ProjectStatus = "CANCELLED"
)

type Project struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string        `gorm:"not null;size:255" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
	Status      ProjectStatus `gorm:"type:project_status;not null;default:'ACTIVE'" json:"status"`
	ManagerID   int           `json:"manager_id"`

	Manager     *User    `gorm:"foreignKey:ManagerID" json:"manager"`
	Tasks       []Task   `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	Sprints     []Sprint `gorm:"foreignKey:ProjectID" json:"sprints,omitempty"`
	TeamMembers []User   `gorm:"foreignKey:CurrentProjectID" json:"team_members,omitempty"`
}

func (ps ProjectStatus) IsValid() bool {
	switch ps {
	case StatusActive, StatusCancelled, StatusCompleted, StatusOnHold:
		return true
	}
	return false
}

func (p *Project) GetID() int {
	return p.ID
}

func (p *Project) GetPKColumnName() string {
	return "id"
}
