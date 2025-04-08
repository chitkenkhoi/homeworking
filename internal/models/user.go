package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	Admin          UserRole = "ADMIN"
	ProjectManager UserRole = "PROJECT_MANAGER"
	TeamMember     UserRole = "TEAM_MEMBER"
)

type User struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email            string   `gorm:"unique;not null;size:255" json:"email"`
	Password         string   `gorm:"not null" json:"-"`
	Role             UserRole `gorm:"type:user_role;not null;default:'TEAM_MEMBER'" json:"role"`
	FirstName        string   `gorm:"size:100" json:"first_name"`
	LastName         string   `gorm:"size:100" json:"last_name"`
	CurrentProjectID *int     `gorm:"index" json:"current_project_id,omitempty"`

	ManagedProjects []Project `gorm:"foreignKey:ManagerID" json:"managed_projects,omitempty"`
	AssignedTasks   []Task    `gorm:"foreignKey:AssigneeID" json:"assigned_tasks,omitempty"`
	CurrentProject  *Project  `gorm:"foreignKey:CurrentProjectID;references:ID" json:"current_project,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = 0
	return nil
}

func (u *User) GetID() int {
	return u.ID
}

func (u *User) GetPKColumnName() string {
	return "id"
}