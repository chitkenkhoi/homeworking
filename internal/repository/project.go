package repository

import (
	"fmt"
	"log/slog"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(project *models.Project) (*models.Project, error)
	Find(filter dto.ProjectFilter) ([]models.Project, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) Create(project *models.Project) (*models.Project, error) {
	if err := r.db.Create(project).Error; err != nil {
		slog.Error("Can not create project", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}
	return project, nil
}

func (r *projectRepository) Find(filter dto.ProjectFilter) ([]models.Project, error) {
	var projects []models.Project
	query := r.db.Model(&models.Project{})

	// Apply filters dynamically
	if filter.ID != nil {
		query = query.Where("id = ?", *filter.ID)
	}
	if filter.Name != nil && *filter.Name != "" {
		// Using LIKE for partial name matching. Use "=" for exact match if needed.
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *filter.Name))
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.ManagerID != nil {
		query = query.Where("manager_id = ?", *filter.ManagerID)
	}
	if filter.StartDateAfter != nil {
		query = query.Where("start_date >= ?", filter.StartDateAfter.Format("2006-01-02"))
	}
	if filter.EndDateBefore != nil {
		query = query.Where("end_date <= ?", filter.EndDateBefore.Format("2006-01-02"))
	}

	query = query.Preload("Manager")

	// Execute the query
	if err := query.Find(&projects).Error; err != nil {
		fmt.Printf("Error finding projects: %v\n", err)
		return nil, fmt.Errorf("database error retrieving projects: %w", err)
	}

	return projects, nil
}
