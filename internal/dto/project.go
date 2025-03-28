package dto

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
)

type CreateProjectRequest struct {
	Name        string               `json:"name" validate:"required,min=2,max=255"`
	Description string               `json:"description" validate:"omitempty,max=65535"`
	StartDate   time.Time            `json:"start_date" validate:"required"`
	EndDate     *time.Time           `json:"end_date" validate:"omitempty,gtfield=StartDate"`
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

type ProjectResponse struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	StartDate   time.Time    `json:"start_date"`
	EndDate     *time.Time   `json:"end_date,omitempty"`
	Status      string       `json:"status"`
	ManagerID   int          `json:"manager_id"`
	TeamMembers []TeamMember `json:"team_members,omitempty"`
}

type TeamMember struct {
	ID        int    `json:"id"`
	Email     string `json:"string"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (pr *ProjectResponse) MapToDto(project *models.Project) {
	pr.Name = project.Name
	pr.Description = project.Description
	pr.StartDate = project.StartDate
	pr.EndDate = project.EndDate
	pr.Status = string(project.Status)
	pr.ManagerID = project.ManagerID
	if project.TeamMembers != nil {
		for _, user := range project.TeamMembers {
			team_member := TeamMember{
				ID:        user.ID,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}
			pr.TeamMembers = append(pr.TeamMembers, team_member)
		}
	}
}

type ProjectFilter struct {
	ID         *int
	Name       *string
	Status     *models.ProjectStatus
	ManagerID  *int
	StartDateAfter  *time.Time 
	EndDateBefore    *time.Time
}