package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	ToDoTask       TaskStatus = "TO_DO"
	InProgressTask TaskStatus = "IN_PROGRESS"
	ReviewTask     TaskStatus = "REVIEW"
	DoneTask       TaskStatus = "DONE"
	BlockedTask    TaskStatus = "BLOCKED"
)

type TaskPriority string

const (
	HighPriority    TaskPriority = "HIGH"
	MediumPriority  TaskPriority = "MEDIUM"
	LowPriority     TaskPriority = "LOW"
	CriticalPriority TaskPriority = "CRITICAL"
)

type Task struct {
	ID        int           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string        `gorm:"not null;size:255" json:"title"`
	Description string        `gorm:"type:text" json:"description"`
	AssigneeID  *int          `gorm:"index" json:"assignee_id"`       
	ProjectID   int           `gorm:"index;not null" json:"project_id"` 
	SprintID    *int          `gorm:"index" json:"sprint_id,omitempty"`
	Status      TaskStatus    `gorm:"type:task_status;not null;default:'TO_DO'" json:"status"`
	Priority    TaskPriority  `gorm:"type:task_priority;not null;default:'MEDIUM'" json:"priority"`
	DueDate     *time.Time    `json:"due_date"`

	Assignee *User `gorm:"foreignKey:AssigneeID;references:ID" json:"assignee,omitempty"`
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"project"`
	Sprint *Sprint `gorm:"foreignKey:SprintID;references:ID" json:"sprint,omitempty"`
}