package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

// CreateProjectRequest represents the request body for creating a new project.
type CreateProjectRequest struct {
	// Name is the name of the project.
	// @example "Website Redesign"
	Name        string               `json:"name" validate:"required,min=2,max=255"`
	// Description is an optional detailed description of the project.
	// @example "Redesign the company website to improve UX."
	Description string               `json:"description" validate:"omitempty,max=65535"`
	// StartDate is the date when the project begins.
	// @example "2025-04-10T00:00:00Z"
	StartDate   time.Time            `json:"start_date" validate:"required"`
	// EndDate is the optional date when the project is expected to end.
	// @example "2025-06-15T00:00:00Z"
	EndDate     *time.Time           `json:"end_date" validate:"omitempty,gtfield=StartDate"`
	// Status is the current status of the project.
	// @example "ACTIVE"
	Status      models.ProjectStatus `json:"status" validate:"omitempty,oneof=ACTIVE ON_HOLD COMPLETED CANCELLED"`
}

func (cpr *CreateProjectRequest) MapToProject(managerID int) *models.Project {
	return &models.Project{
		Name:        cpr.Name,
		Description: cpr.Description,
		StartDate:   cpr.StartDate,
		EndDate:     cpr.EndDate,
		Status:      cpr.Status,
		ManagerID:   managerID,
	}
}

// ProjectResponse represents the response body for project details.
type ProjectResponse struct {
	// ID is the unique identifier of the project.
	// @example 1
	ID              int          `json:"id"`
	// Name is the name of the project.
	// @example "Website Redesign"
	Name            string       `json:"name"`
	// Description is the detailed description of the project.
	// @example "Redesign the company website to improve UX."
	Description     string       `json:"description"`
	// StartDate is the date when the project started.
	// @example "2025-04-10T00:00:00Z"
	StartDate       time.Time    `json:"start_date"`
	// EndDate is the date when the project is expected to end (optional).
	// @example "2025-06-15T00:00:00Z"
	EndDate         *time.Time   `json:"end_date,omitempty"`
	// Status is the current status of the project.
	// @example "ACTIVE"
	Status          string       `json:"status"`
	// ManagerID is the ID of the project manager.
	// @example 42
	ManagerID       int          `json:"manager_id"`
	// TeamMembers is the list of team members assigned to the project (optional).
	TeamMembers     []TeamMember `json:"team_members,omitempty"`
	// TeamMemberCount is the total number of team members (optional).
	// @example 3
	TeamMemberCount *int         `json:"team_member_count,omitempty"`
}

// TeamMember represents a team member assigned to a project.
type TeamMember struct {
	// ID is the unique identifier of the team member.
	// @example 101
	ID        int    `json:"id"`
	// Email is the email address of the team member.
	// @example "john.doe@example.com"
	Email     string `json:"email"`
	// FirstName is the first name of the team member.
	// @example "John"
	FirstName string `json:"first_name"`
	// LastName is the last name of the team member.
	// @example "Doe"
	LastName  string `json:"last_name"`
}

func MapToProjectDto(project *models.Project) *ProjectResponse {
	pr := &ProjectResponse{}
	pr.ID = project.ID
	pr.Name = project.Name
	pr.Description = project.Description
	pr.StartDate = project.StartDate
	pr.EndDate = project.EndDate
	pr.Status = string(project.Status)
	pr.ManagerID = project.ManagerID
	if len(project.TeamMembers) == 0 {
		return pr
	}
	count := len(project.TeamMembers)
	pr.TeamMemberCount = &count

	pr.TeamMembers = make([]TeamMember, count)
	for i, user := range project.TeamMembers {
		pr.TeamMembers[i] = TeamMember{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}
	}

	return pr
}

// ProjectFilter represents filtering options for querying projects.
type ProjectFilter struct {
	// ID is the optional project ID to filter by.
	// @example 1
	ID             *int
	// Name is the optional project name to filter by.
	// @example "Website Redesign"
	Name           *string
	// Status is the optional project status to filter by.
	// @example "ACTIVE"
	Status         *models.ProjectStatus
	// ManagerID is the optional manager ID to filter by.
	// @example 42
	ManagerID      *int
	// StartDateAfter is the optional start date to filter projects starting after.
	// @example "2025-04-01T00:00:00Z"
	StartDateAfter *time.Time
	// EndDateBefore is the optional end date to filter projects ending before.
	// @example "2025-07-01T00:00:00Z"
	EndDateBefore  *time.Time
}

func MapToProjectDtoSlice(projects []*models.Project) []ProjectResponse {
	prs := make([]ProjectResponse, 0, len(projects))
	for _, project := range projects {
		prs = append(prs, ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			StartDate:   project.StartDate,
			EndDate:     project.EndDate,
			Status:      string(project.Status),
			ManagerID:   project.ManagerID,
		})
	}
	return prs
}

// AddTeamMembersRequest represents the request body for adding team members to a project.
type AddTeamMembersRequest struct {
	// UserIDs is the list of user IDs to add as team members.
	// @example [101, 102, 103]
	UserIDs []int `json:"userIds" validate:"required,min=1,dive,gt=0"`
}

// UpdateProjectRequest represents the request body for updating an existing project.
type UpdateProjectRequest struct {
	// Name is the optional new name of the project.
	// @example "Website Redesign v2"
	Name        *string               `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	// Description is the optional new description of the project.
	// @example "Updated redesign with new features."
	Description *string               `json:"description,omitempty" validate:"omitempty,max=65535"`
	// EndDate is the optional new end date of the project.
	// @example "2025-07-01T00:00:00Z"
	EndDate     *time.Time            `json:"end_date,omitempty" validate:"omitempty"`
	// Status is the optional new status of the project.
	// @example "ON_HOLD"
	Status      *models.ProjectStatus `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE ON_HOLD COMPLETED CANCELLED"`
}