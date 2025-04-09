package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

// CreateSprintRequest represents the request body for creating a new sprint.
type CreateSprintRequest struct {
	// Name is the name of the sprint.
	Name      string    `json:"name" validate:"required,min=2" example:"Sprint 1"`
	// StartDate is the date when the sprint begins.
	StartDate time.Time `json:"start_date" validate:"required" example:"2025-04-15T00:00:00Z"`
	// EndDate is the date when the sprint ends.
	EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate" example:"2025-04-30T00:00:00Z"`
	// ProjectID is the ID of the project this sprint belongs to.
	ProjectID int       `json:"project_id" validate:"required,min=1" example:"1"`
	// Goal is the objective or goal of the sprint.
	Goal      string    `json:"goal" validate:"required,min=5" example:"Complete initial UI design"`
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

// TaskInSprintResponse represents a task within a sprint in the response.
type TaskInSprintResponse struct {
	// ID is the unique identifier of the task.
	ID       int                 `json:"id" example:"101"`
	// Title is the title or description of the task.
	Title    string              `json:"task" example:"Design homepage layout"`
	// Status is the current status of the task.
	Status   models.TaskStatus   `json:"status" example:"IN_PROGRESS"`
	// Priority is the priority level of the task.
	Priority models.TaskPriority `json:"priority" example:"HIGH"`
	// DueDate is the optional due date of the task.
	DueDate  *time.Time          `json:"due_date" example:"2025-04-25T00:00:00Z"`
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

// SprintResponse represents the response body for sprint details.
type SprintResponse struct {
	// ID is the unique identifier of the sprint.
	ID          int                    `json:"id" example:"1"`
	// Name is the name of the sprint.
	Name        string                 `json:"name" example:"Sprint 1"`
	// StartDate is the date when the sprint started.
	StartDate   time.Time              `json:"start_date" example:"2025-04-15T00:00:00Z"`
	// EndDate is the date when the sprint ends.
	EndDate     time.Time              `json:"end_date" example:"2025-04-30T00:00:00Z"`
	// ProjectID is the ID of the project this sprint belongs to.
	ProjectID   int                    `json:"project_id" example:"1"`
	// ProjectName is the optional name of the associated project.
	ProjectName *string                `json:"project_name,omitempty" example:"Website Redesign"`
	// Goal is the objective or goal of the sprint.
	Goal        string                 `json:"goal" example:"Complete initial UI design"`
	// Tasks is the list of tasks within the sprint (optional).
	Tasks       []TaskInSprintResponse `json:"tasks,omitempty"`
	// TaskCount is the total number of tasks in the sprint (optional).
	TaskCount   *int                   `json:"task_count,omitempty" example:"3"`
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

func MapToSprintResponseSlice(sprints []*models.Sprint) []SprintResponse {
	if sprints == nil {
		return []SprintResponse{}
	}
	responses := make([]SprintResponse, len(sprints))

	for i := range sprints {
		sprints[i].Tasks = nil
		responses[i] = *MapToSprintResponse(sprints[i])
	}

	return responses
}

// SprintFilter represents filtering options for querying sprints.
type SprintFilter struct {
	// ID is the optional sprint ID to filter by.
	ID             *int
	// Name is the optional sprint name to filter by.
	Name           *string
	// ProjectID is the optional project ID to filter by.
	ProjectID      *int
	// StartDateAfter is the optional start date to filter sprints starting after.
	StartDateAfter *time.Time
	// EndDateBefore is the optional end date to filter sprints ending before.
	EndDateBefore  *time.Time
}

// UpdateSprintRequest represents the request body for updating an existing sprint.
type UpdateSprintRequest struct {
	// Name is the optional new name of the sprint.
	Name      *string    `json:"name,omitempty" validate:"omitempty,min=2" example:"Sprint 1 - Revised"`
	// StartDate is the optional new start date of the sprint.
	StartDate *time.Time `json:"start_date,omitempty" validate:"omitempty" example:"2025-04-20T00:00:00Z"`
	// EndDate is the optional new end date of the sprint.
	EndDate   *time.Time `json:"end_date,omitempty" validate:"omitempty,gtfield=StartDate" example:"2025-05-05T00:00:00Z"`
	// Goal is the optional new goal of the sprint.
	Goal      *string    `json:"goal,omitempty" validate:"omitempty,min=5" example:"Finalize UI and start backend integration"`
}