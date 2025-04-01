package repository

import (
	"context"
	"errors"
	"fmt"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type SprintRepository interface {
	Create(ctx context.Context, sprint *models.Sprint) (*models.Sprint, error)
	FindByID(ctx context.Context, id int) (*models.Sprint, error)
	Find(ctx context.Context, filter dto.SprintFilter) ([]models.Sprint, error)
	Delete(ctx context.Context, id int) error
}

type sprintRepository struct {
	db  *gorm.DB
	cfg config.DateTimeConfig
}

func NewSprintRepository(db *gorm.DB, cfg config.DateTimeConfig) SprintRepository {
	return &sprintRepository{
		db:  db,
		cfg: cfg,
	}
}

func (r *sprintRepository) Create(ctx context.Context, sprint *models.Sprint) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Create",
	)

	logger.Debug("Starting create sprint process")

	if err := r.db.Create(sprint).Error; err != nil {
		logger.Error("Can not create project", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}

	logger.Debug("Create sprint successfully", "sprint", *sprint)

	return sprint, nil
}

func (r *sprintRepository) FindByID(ctx context.Context, id int) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "FindByID",
	)

	logger.Debug("Starting get sprint process")
	var sprint models.Sprint
	err := r.db.Preload("Tasks").Preload("Project").First(&sprint, id).Error
	if err != nil {
		logger.Error("Failed to get sprint", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, structs.ErrSprintNotExist
		}
		return nil, structs.ErrDatabaseFail
	}

	logger.Debug("Successfully retrieved sprint with tasks", "sprintID", sprint.ID, "taskCount", len(sprint.Tasks))
	return &sprint, nil

}

func (r *sprintRepository) Find(ctx context.Context, filter dto.SprintFilter) ([]models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Find",
	)
	var sprints []models.Sprint
	query := r.db.Model(&models.Sprint{})

	if filter.ID != nil {
		query = query.Where("id = ?", *filter.ID)
		logger.Debug("Sprint ID is in the filter", "sprint_id", *filter.ID)
	}
	if filter.Name != nil && *filter.Name != "" {
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *filter.Name))
		logger.Debug("Sprint name is in the filter", "sprint_name", *filter.Name)
	}
	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", *filter.ProjectID)
		logger.Debug("Project ID is in the filter", "project_id", *filter.ProjectID)
	}
	if filter.StartDateAfter != nil {
		query = query.Where("start_date >= ?", filter.StartDateAfter.Format(r.cfg.Format))
		logger.Debug("Start date is in the filter", "start_date_after", filter.StartDateAfter.Format(r.cfg.Format))
	}
	if filter.EndDateBefore != nil {
		query = query.Where("end_date <= ?", filter.EndDateBefore.Format(r.cfg.Format))
		logger.Debug("End date is in the filter", "end_date_after", filter.EndDateBefore.Format(r.cfg.Format))
	}

	if err := query.Find(&sprints).Error; err != nil {
		logger.Error("Error finding sprints", "error", err)
		return nil, fmt.Errorf("database error retrieving sprints: %w", err)
	}
	logger.Debug("Found sprints", "sprints", sprints)
	return sprints, nil
}

func (r *sprintRepository) Delete(ctx context.Context, id int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Delete",
		"sprint_id", id,
	)

	logger.Debug("Starting delete sprint process")

	tx := r.db.Delete(&models.Sprint{}, id)
	err := tx.Error
	if err != nil {
		logger.Error("Internal database fail", "error", err)
		return structs.ErrDatabaseFail
	}

	if tx.RowsAffected == 0 {
		logger.Error("Sprint with this id does not exist")
		return structs.ErrProjectNotExist
	}

	logger.Info("Project is deleted")

	return nil
}
