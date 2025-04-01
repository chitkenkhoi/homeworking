package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) (*models.Project, error)
	Find(ctx context.Context, filter dto.ProjectFilter) ([]models.Project, error)
	FindByID(ctx context.Context, id int) (*models.Project, error)
	Update(ctx context.Context, id int, updateMap *map[string]any) error
	Delete(ctx context.Context, id int) error
}

type projectRepository struct {
	db  *gorm.DB
	cfg config.DateTimeConfig
}

func NewProjectRepository(db *gorm.DB, cfg config.DateTimeConfig) ProjectRepository {
	return &projectRepository{
		db:  db,
		cfg: cfg,
	}
}

func (r *projectRepository) Create(ctx context.Context, project *models.Project) (*models.Project, error) {
	if err := r.db.Create(project).Error; err != nil {
		slog.Error("Can not create project", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}
	return project, nil
}

func (r *projectRepository) Find(ctx context.Context, filter dto.ProjectFilter) ([]models.Project, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectRepository",
		"method", "Find",
	)

	var projects []models.Project
	query := r.db.Model(&models.Project{})

	if filter.ID != nil {
		query = query.Where("id = ?", *filter.ID)
		logger.Debug("Project ID is in the filter", "project_id", *filter.ID)
	}
	if filter.Name != nil && *filter.Name != "" {
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *filter.Name))
		logger.Debug("Project name is in the filter", "project_name", *filter.Name)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
		logger.Debug("Project status is in the filter", "project_status", *filter.Status)
	}
	if filter.ManagerID != nil {
		query = query.Where("manager_id = ?", *filter.ManagerID)
		logger.Debug("Manager ID is in the filter", "project_id", *filter.ManagerID)
	}
	if filter.StartDateAfter != nil {
		query = query.Where("start_date >= ?", filter.StartDateAfter.Format(r.cfg.Format))
		logger.Debug("Start date is in the filter", "start_date_after", filter.StartDateAfter.Format(r.cfg.Format))
	}
	if filter.EndDateBefore != nil {
		query = query.Where("end_date <= ?", filter.EndDateBefore.Format(r.cfg.Format))
		logger.Debug("End date is in the filter", "end_date_after", filter.EndDateBefore.Format(r.cfg.Format))
	}

	// Execute the query
	if err := query.Find(&projects).Error; err != nil {
		logger.Error("Error finding projects", "error", err)
		return nil, fmt.Errorf("database error retrieving projects: %w", err)
	}
	logger.Debug("Found projects", "projects", projects)

	return projects, nil
}

func (r *projectRepository) FindByID(ctx context.Context, id int) (*models.Project, error) {
	var project models.Project
	err := r.db.First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, structs.ErrProjectNotExist
		}
		slog.Error("Internal database failed", "err", err)
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) Update(ctx context.Context, id int, updateMap *map[string]any) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectRepository",
		"method", "Update",
		"project_id", id,
	)

	logger.Debug("Starting update project process")

	if err := r.db.Model(&models.Project{}).Where("id = ?", id).Updates(*updateMap).Error; err != nil {
		logger.Error("Failed to update project", "error", err)
		return fmt.Errorf("failed to update project: %w", err)
	}

	logger.Info("Succesfully updated")

	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectRepository",
		"method", "Delete",
		"project_id", id,
	)

	logger.Debug("Starting delete project process")

	tx := r.db.Delete(&models.Project{}, id)
	err := tx.Error
	if err != nil {
		logger.Error("Internal database fail", "error", err)
		return structs.ErrDatabaseFail
	}

	if tx.RowsAffected == 0 {
		logger.Error("Project with this id does not exist")
		return structs.ErrProjectNotExist
	}

	logger.Info("Project is deleted")

	return nil
}
