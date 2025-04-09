package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

// CreateTaskRequest represents the request body for creating a new task.
type CreateTaskRequest struct {
	// Title is the title of the task.
	Title       string              `json:"title" validate:"required,min=2,max=255" example:"Implement login API"`
	// Description is an optional detailed description of the task.
	Description string              `json:"description" validate:"omitempty,max=65535" example:"Create a RESTful endpoint for user authentication."`
	// SprintID is the ID of the sprint this task belongs to.
	SprintID    int                 `json:"sprint_id" validate:"required,min=1" example:"1"`
	// Status is the current status of the task.
	Status      models.TaskStatus   `json:"status" validate:"omitempty,oneof=TO_DO IN_PROGRESS REVIEW DONE BLOCKED" example:"TO_DO"`
	// Priority is the priority level of the task.
	Priority    models.TaskPriority `json:"priority" validate:"omitempty,oneof=HIGH MEDIUM LOW CRITICAL" example:"HIGH"`
	// DueDate is the optional due date of the task.
	DueDate     *time.Time          `json:"due_date,omitempty" validate:"omitempty" example:"2025-04-20T00:00:00Z"`
}

func (ctr *CreateTaskRequest) MapToTask() *models.Task {
	return &models.Task{
		Title:       ctr.Title,
		Description: ctr.Description,
		ProjectID:   0,
		SprintID:    ctr.SprintID,
		Status:      ctr.Status,
		Priority:    ctr.Priority,
		DueDate:     ctr.DueDate,
	}
}

// TaskResponse represents the response body for detailed task information.
type TaskResponse struct {
	// ID is the unique identifier of the task.
	ID                int                 `json:"id" example:"101"`
	// Title is the title of the task.
	Title             string              `json:"title" example:"Implement login API"`
	// Description is the detailed description of the task.
	Description       string              `json:"description" example:"Create a RESTful endpoint for user authentication."`
	// AssigneeID is the optional ID of the user assigned to the task.
	AssigneeID        *int                `json:"assignee_id,omitempty" example:"42"`
	// AssigneeFirstName is the optional first name of the assignee.
	AssigneeFirstName *string             `json:"assignee_first_name,omitempty" example:"John"`
	// AssigneeLastName is the optional last name of the assignee.
	AssigneeLastName  *string             `json:"assignee_last_name,omitempty" example:"Doe"`
	// ProjectID is the ID of the project this task belongs to.
	ProjectID         int                 `json:"project_id" example:"1"`
	// ProjectName is the name of the project this task belongs to.
	ProjectName       string              `json:"project_name" example:"Website Redesign"`
	// SprintID is the ID of the sprint this task belongs to.
	SprintID          int                 `json:"sprint_id" example:"1"`
	// SprintName is the name of the sprint this task belongs to.
	SprintName        string              `json:"sprint_name" example:"Sprint 1"`
	// Status is the current status of the task.
	Status            models.TaskStatus   `json:"status" example:"IN_PROGRESS"`
	// Priority is the priority level of the task.
	Priority          models.TaskPriority `json:"priority" example:"HIGH"`
	// DueDate is the optional due date of the task.
	DueDate           *time.Time          `json:"due_date,omitempty" example:"2025-04-20T00:00:00Z"`
}

func MapToTaskResponse(task *models.Task) *TaskResponse {
	response := &TaskResponse{}
	response.ID = task.ID
	response.Title = task.Title
	response.Description = task.Description

	if task.AssigneeID != nil {
		response.AssigneeFirstName = &task.Assignee.FirstName
		response.AssigneeLastName = &task.Assignee.LastName
	}

	response.ProjectID = task.ProjectID
	response.SprintID = task.SprintID

	if task.Project != nil {
		response.ProjectName = task.Project.Name
	} else if task.Sprint != nil {
		response.ProjectName = task.Sprint.Project.Name
	} else {
		response.ProjectName = "" //something wrong if reach this
	}

	if task.Sprint != nil {
		response.SprintName = task.Sprint.Name
	} else {
		response.SprintName = "" //something wrong if reach this
	}

	response.Status = task.Status
	response.Priority = task.Priority
	response.DueDate = task.DueDate

	return response
}

// TaskInSliceResponse represents a task in a list response.
type TaskInSliceResponse struct {
	// ID is the unique identifier of the task.
	ID                int                 `json:"id" example:"101"`
	// Title is the title of the task.
	Title             string              `json:"title" example:"Implement login API"`
	// Description is the detailed description of the task.
	Description       string              `json:"description" example:"Create a RESTful endpoint for user authentication."`
	// SprintID is the ID of the sprint this task belongs to.
	SprintID          int                 `json:"sprint_id" example:"1"`
	// ProjectID is the ID of the project this task belongs to.
	ProjectID         int                 `json:"project_id" example:"1"`
	// AssigneeID is the optional ID of the user assigned to the task.
	AssigneeID        *int                `json:"assignee_id,omitempty" example:"42"`
	// AssigneeFirstName is the optional first name of the assignee.
	AssigneeFirstName *string             `json:"assignee_first_name,omitempty" example:"John"`
	// AssigneeLastName is the optional last name of the assignee.
	AssigneeLastName  *string             `json:"assignee_last_name,omitempty" example:"Doe"`
	// Status is the current status of the task.
	Status            models.TaskStatus   `json:"status" example:"IN_PROGRESS"`
	// Priority is the priority level of the task.
	Priority          models.TaskPriority `json:"priority" example:"HIGH"`
	// DueDate is the optional due date of the task.
	DueDate           *time.Time          `json:"due_date,omitempty" example:"2025-04-20T00:00:00Z"`
}

func MapToSliceOfTaskResponse(tasks []*models.Task) []TaskInSliceResponse {
	if tasks == nil {
		return []TaskInSliceResponse{}
	}

	res := make([]TaskInSliceResponse, len(tasks))

	for i, task := range tasks {
		res[i].ID = task.ID
		res[i].Title = task.Title
		res[i].Description = task.Description
		res[i].SprintID = task.SprintID
		res[i].Status = task.Status
		res[i].Priority = task.Priority
		res[i].DueDate = task.DueDate
		res[i].AssigneeID = task.AssigneeID
		res[i].ProjectID = task.ProjectID

		if task.Assignee != nil {
			res[i].AssigneeFirstName = &task.Assignee.FirstName
			res[i].AssigneeLastName = &task.Assignee.LastName
		}
	}
	return res
}

// UpdateTaskRequest represents the request body for updating an existing task.
type UpdateTaskRequest struct {
	// Title is the optional new title of the task.
	Title       *string              `json:"title" validate:"omitempty,min=2,max=255" example:"Update login API"`
	// Description is the optional new description of the task.
	Description *string              `json:"description" validate:"omitempty,max=65535" example:"Modify endpoint to include JWT."`
	// Status is the optional new status of the task.
	Status      *models.TaskStatus   `json:"status" validate:"omitempty,oneof=TO_DO IN_PROGRESS REVIEW DONE BLOCKED" example:"REVIEW"`
	// Priority is the optional new priority level of the task.
	Priority    *models.TaskPriority `json:"priority" validate:"omitempty,oneof=HIGH MEDIUM LOW CRITICAL" example:"MEDIUM"`
	// DueDate is the optional new due date of the task.
	DueDate     *time.Time           `json:"due_date,omitempty" validate:"omitempty" example:"2025-04-25T00:00:00Z"`
}

// TaskFilter represents filtering options for querying tasks.
type TaskFilter struct {
	// ID is the optional task ID to filter by.
	ID            *int
	// Title is the optional task title to filter by.
	Title         *string
	// Status is the optional task status to filter by.
	Status        *models.TaskStatus
	// Priority is the optional task priority to filter by.
	Priority      *models.TaskPriority
	// DueDateBefore is the optional due date to filter tasks due before.
	DueDateBefore *time.Time
}