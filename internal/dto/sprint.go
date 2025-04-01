package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

type CreateSprintRequest struct {
	Name      string    `json:"name" validate:"required,min=2"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	ProjectID int       `json:"project_id" validate:"required,min=1"`
	Goal      string    `json:"goal" validate:"rqeuired,min=5"`
}

func (csr *CreateSprintRequest) MapToSprint() *models.Sprint {
	return &models.Sprint{
		Name:      csr.Name,
		StartDate: csr.StartDate,
		EndDate:   csr.EndDate,
		ProjectID: csr.ProjectID,
		Goal:      csr.Goal,
	}
}

type TaskInSprintResponse struct {
	ID       int                 `json:"id"`
	Title    string              `json:"task"`
	Status   models.TaskStatus   `json:"status"`
	Priority models.TaskPriority `json:"priority"`
	DueDate  *time.Time          `json:"due_date"`
}

func MapToTaskInSprintResponse(task *models.Task) *TaskInSprintResponse {
	return &TaskInSprintResponse{
		ID:       task.ID,
		Title:    task.Title,
		Status:   task.Status,
		Priority: task.Priority,
		DueDate:  task.DueDate,
	}
}

type SprintResponse struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	StartDate   time.Time              `json:"start_date"`
	EndDate     time.Time              `json:"end_date"`
	ProjectID   int                    `json:"project_id"`
	ProjectName *string                `json:"project_name,omitempty"`
	Goal        string                 `json:"goal"`
	Tasks       []TaskInSprintResponse `json:"tasks,omitempty"`
	TaskCount   *int                   `json:"task_count,omitempty"`
}

func MapToSprintResponse(sprint *models.Sprint) *SprintResponse {
	sr := &SprintResponse{}
	sr.ID = sprint.ID
	sr.Name = sprint.Name
	sr.StartDate = sprint.StartDate
	sr.EndDate = sprint.EndDate
	sr.ProjectID = sprint.ProjectID

	if sprint.Project != nil {
		sr.ProjectName = &sprint.Project.Name
	}

	sr.Goal = sprint.Goal
	if len(sprint.Tasks) == 0 {
		return sr
	}

	numTasks := len(sprint.Tasks)
	sr.TaskCount = &numTasks
	sr.Tasks = make([]TaskInSprintResponse, numTasks)

	for i := range sprint.Tasks {
		sr.Tasks[i] = *MapToTaskInSprintResponse(&sprint.Tasks[i])
	}

	return sr
}

func MapToSprintResponseSlice(sprints []models.Sprint) []SprintResponse {
	responses := make([]SprintResponse, len(sprints))

	for i := range sprints {
		sprints[i].Tasks = nil
		responses[i] = *MapToSprintResponse(&sprints[i])
	}

	return responses
}

type SprintFilter struct {
	ID             *int
	Name           *string
	ProjectID      *int
	StartDateAfter *time.Time
	EndDateBefore  *time.Time
}
