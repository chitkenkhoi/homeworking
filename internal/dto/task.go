package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

type CreateTaskRequest struct {
	Title       string              `json:"tittle"             validate:"required,min=2,max=255"`
	Description string              `json:"description"        validate:"omitempty,max=65535"`
	SprintID    int                 `json:"sprint_id"          validate:"required,min=1"`
	Status      models.TaskStatus   `json:"status"             validate:"omitempty,oneof=TO_DO IN_PROGRESS REVIEW DONE BLOCKED"`
	Priority    models.TaskPriority `json:"priority"           validate:"omitempty,oneof=HIGH MEDIUM LOW CRITICAL"`
	DueDate     *time.Time          `json:"due_date,omitempty" validate:"omitempty"`
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

type TaskResponse struct {
	ID                int                 `json:"id"`
	Title             string              `json:"title"`
	Description       string              `json:"description"`
	AssigneeID        *int                `json:"assignee_id,omitempty"`
	AssigneeFirstName *string             `json:"assignee_first_name,omitempty"`
	AssigneeLastName  *string             `json:"assignee_last_name,omitempty"`
	ProjectID         int                 `json:"project_id"`
	ProjectName       string              `json:"project_name"`
	SprintID          int                 `json:"sprint_id"`
	SprintName        string              `json:"sprint_name"`
	Status            models.TaskStatus   `json:"status"`
	Priority          models.TaskPriority `json:"priority"`
	DueDate           *time.Time          `json:"due_date,omitempty"`
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
		response.ProjectName = "" //some thing wrong if reach this
	}

	if task.Sprint != nil {
		response.SprintName = task.Sprint.Name
	} else {
		response.SprintName = "" //some thing wrong if reach this
	}

	response.Status = task.Status
	response.Priority = task.Priority
	response.DueDate = task.DueDate

	return response
}

type TaskInSliceResponse struct {
	ID                int                 `json:"id"`
	Title             string              `json:"tittle"`
	Description       string              `json:"description"`
	SprintID          int                 `json:"sprint_id"`
	ProjectID         int                 `json:"project_id"`
	AssigneeID        *int                `json:"assignee_id,omitempty"`
	AssigneeFirstName *string             `json:"assignee_first_name,omitempty"`
	AssigneeLastName  *string             `json:"assignee_last_name,omitempty"`
	Status            models.TaskStatus   `json:"status"`
	Priority          models.TaskPriority `json:"priority"`
	DueDate           *time.Time          `json:"due_date,omitempty"`
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

type UpdateTaskRequest struct {
	Title       *string              `json:"title"              validate:"omitempty,min=2,max=255"`
	Description *string              `json:"description"        validate:"omitempty,max=65535"`
	Status      *models.TaskStatus   `json:"status"             validate:"omitempty,oneof=TO_DO IN_PROGRESS REVIEW DONE BLOCKED"`
	Priority    *models.TaskPriority `json:"priority"           validate:"omitempty,oneof=HIGH MEDIUM LOW CRITICAL"`
	DueDate     *time.Time           `json:"due_date,omitempty" validate:"omitempty"`
}

type TaskFilter struct {
	ID            *int
	Title         *string
	Status        *models.TaskStatus
	Priority      *models.TaskPriority
	DueDateBefore *time.Time
}
