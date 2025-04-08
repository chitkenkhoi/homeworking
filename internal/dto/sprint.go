package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

// CreateSprintRequest represents the request body for creating a new sprint.
type CreateSprintRequest struct {
	// Name is the name of the sprint.
	// @example "Sprint 1"
	Name      string    `json:"name"       validate:"required,min=2"`
	// StartDate is the date when the sprint begins.
	// @example "2025-04-15T00:00:00Z"
	StartDate time.Time `json:"start_date" validate:"required"`
	// EndDate is the date when the sprint ends.
	// @example "2025-04-30T00:00:00Z"
	EndDate   time.Time `json:"end_date"   validate:"required,gtfield=StartDate"`
	// ProjectID is the ID of the project this sprint belongs to.
	// @example 1
	ProjectID int       `json:"project_id" validate:"required,min=1"`
	// Goal is the objective or goal of the sprint.
	// @example "Complete initial UI design"
	Goal      string    `json:"goal"       validate:"required,min=5"`
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
	// @example 101
	ID       int                 `json:"id"`
	// Title is the title or description of the task.
	// @example "Design homepage layout"
	Title    string              `json:"task"`
	// Status is the current status of the task.
	// @example "IN_PROGRESS"
	Status   models.TaskStatus   `json:"status"`
	// Priority is the priority level of the task.
	// @example "HIGH"
	Priority models.TaskPriority `json:"priority"`
	// DueDate is the optional due date of the task.
	// @example "2025-04-25T00:00:00Z"
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

// SprintResponse represents the response body for sprint details.
type SprintResponse struct {
	// ID is the unique identifier of the sprint.
	// @example 1
	ID          int                    `json:"id"`
	// Name is the name of the sprint.
	// @example "Sprint 1"
	Name        string                 `json:"name"`
	// StartDate is the date when the sprint started.
	// @example "2025-04-15T00:00:00Z"
	StartDate   time.Time              `json:"start_date"`
	// EndDate is the date when the sprint ends.
	// @example "2025-04-30T00:00:00Z"
	EndDate     time.Time              `json:"end_date"`
	// ProjectID is the ID of the project this sprint belongs to.
	// @example 1
	ProjectID   int                    `json:"project_id"`
	// ProjectName is the optional name of the associated project.
	// @example "Website Redesign"
	ProjectName *string                `json:"project_name,omitempty"`
	// Goal is the objective or goal of the sprint.
	// @example "Complete initial UI design"
	Goal        string                 `json:"goal"`
	// Tasks is the list of tasks within the sprint (optional).
	Tasks       []TaskInSprintResponse `json:"tasks,omitempty"`
	// TaskCount is the total number of tasks in the sprint (optional).
	// @example 3
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
	// @example 1
	ID             *int
	// Name is the optional sprint name to filter by.
	// @example "Sprint 1"
	Name           *string
	// ProjectID is the optional project ID to filter by.
	// @example 1
	ProjectID      *int
	// StartDateAfter is the optional start date to filter sprints starting after.
	// @example "2025-04-01T00:00:00Z"
	StartDateAfter *time.Time
	// EndDateBefore is the optional end date to filter sprints ending before.
	// @example "2025-05-01T00:00:00Z"
	EndDateBefore  *time.Time
}

// UpdateSprintRequest represents the request body for updating an existing sprint.
type UpdateSprintRequest struct {
	// Name is the optional new name of the sprint.
	// @example "Sprint 1 - Revised"
	Name      *string    `json:"name,omitempty"       validate:"omitempty,min=2"`
	// StartDate is the optional new start date of the sprint.
	// @example "2025-04-20T00:00:00Z"
	StartDate *time.Time `json:"start_date,omitempty" validate:"omitempty"`
	// EndDate is the optional new end date of the sprint.
	// @example "2025-05-05T00:00:00Z"
	EndDate   *time.Time `json:"end_date,omitempty"   validate:"omitempty,gtfield=StartDate"`
	// Goal is the optional new goal of the sprint.
	// @example "Finalize UI and start backend integration"
	Goal      *string    `json:"goal,omitempty"       validate:"omitempty,min=5"`
}