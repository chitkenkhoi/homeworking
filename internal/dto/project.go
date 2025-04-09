package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

// CreateProjectRequest represents the request body for creating a new project.
type CreateProjectRequest struct {
	// Name is the name of the project.
	Name        string               `json:"name" validate:"required,min=2,max=255" example:"Website Redesign"`
	// Description is an optional detailed description of the project.
	Description string               `json:"description" validate:"omitempty,max=65535" example:"Redesign the company website to improve UX."`
	// StartDate is the date when the project begins.
	StartDate   time.Time            `json:"start_date" validate:"required" example:"2025-04-10T00:00:00Z"`
	// EndDate is the optional date when the project is expected to end.
	EndDate     *time.Time           `json:"end_date" validate:"omitempty,gtfield=StartDate" example:"2025-06-15T00:00:00Z"`
	// Status is the current status of the project.
	Status      models.ProjectStatus `json:"status" validate:"omitempty,oneof=ACTIVE ON_HOLD COMPLETED CANCELLED" example:"ACTIVE"`
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
	ID              int          `json:"id" example:"1"`
	// Name is the name of the project.
	Name            string       `json:"name" example:"Website Redesign"`
	// Description is the detailed description of the project.
	Description     string       `json:"description" example:"Redesign the company website to improve UX."`
	// StartDate is the date when the project started.
	StartDate       time.Time    `json:"start_date" example:"2025-04-10T00:00:00Z"`
	// EndDate is the date when the project is expected to end (optional).
	EndDate         *time.Time   `json:"end_date,omitempty" example:"2025-06-15T00:00:00Z"`
	// Status is the current status of the project.
	Status          string       `json:"status" example:"ACTIVE"`
	// ManagerID is the ID of the project manager.
	ManagerID       int          `json:"manager_id" example:"42"`
	// TeamMembers is the list of team members assigned to the project (optional).
	TeamMembers     []TeamMember `json:"team_members,omitempty"`
	// TeamMemberCount is the total number of team members (optional).
	TeamMemberCount *int         `json:"team_member_count,omitempty" example:"3"`
}

// TeamMember represents a team member assigned to a project.
type TeamMember struct {
	// ID is the unique identifier of the team member.
	ID        int    `json:"id" example:"101"`
	// Email is the email address of the team member.
	Email     string `json:"email" example:"john.doe@example.com"`
	// FirstName is the first name of the team member.
	FirstName string `json:"first_name" example:"John"`
	// LastName is the last name of the team member.
	LastName  string `json:"last_name" example:"Doe"`
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
	ID             *int
	// Name is the optional project name to filter by.
	Name           *string
	// Status is the optional project status to filter by.
	Status         *models.ProjectStatus
	// ManagerID is the optional manager ID to filter by.
	ManagerID      *int
	// StartDateAfter is the optional start date to filter projects starting after.
	StartDateAfter *time.Time
	// EndDateBefore is the optional end date to filter projects ending before.
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
	UserIDs []int `json:"userIds" validate:"required,min=1,dive,gt=0" example:"[101, 102, 103]"`
}

// UpdateProjectRequest represents the request body for updating an existing project.
type UpdateProjectRequest struct {
	// Name is the optional new name of the project.
	Name        *string               `json:"name,omitempty" validate:"omitempty,min=2,max=255" example:"Website Redesign v2"`
	// Description is the optional new description of the project.
	Description *string               `json:"description,omitempty" validate:"omitempty,max=65535" example:"Updated redesign with new features."`
	// EndDate is the optional new end date of the project.
	EndDate     *time.Time            `json:"end_date,omitempty" validate:"omitempty" example:"2025-07-01T00:00:00Z"`
	// Status is the optional new status of the project.
	Status      *models.ProjectStatus `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE ON_HOLD COMPLETED CANCELLED" example:"ON_HOLD"`
}