package repository

import (
	"context"
	"errors"
	"fmt"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/query"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type SprintRepository interface {
	Create(ctx context.Context, sprint *models.Sprint) (*models.Sprint, error)
	FindByID(ctx context.Context, id int) (*models.Sprint, error)
	Find(ctx context.Context, filter *dto.SprintFilter) ([]*models.Sprint, error)
	Update(ctx context.Context, id int, updateMap map[string]any) error
	Delete(ctx context.Context, id int) error
}

type sprintRepository struct {
	db  *gorm.DB
	cfg config.DateTimeConfig
	q   *query.Query
}

func NewSprintRepository(db *gorm.DB, cfg config.DateTimeConfig) SprintRepository {
	return &sprintRepository{
		db:  db,
		cfg: cfg,
		q:   query.Use(db),
	}
}

func (r *sprintRepository) Create(ctx context.Context, sprint *models.Sprint) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Create",
	)
	logger.Debug("Starting create sprint process", "sprint_name", sprint.Name, "project_id", sprint.ProjectID)

	err := r.q.Sprint.WithContext(ctx).Create(sprint)
	if err != nil {
		logger.Error("Failed to create sprint", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}

	logger.Info("Successfully created sprint", "sprint_id", sprint.ID)
	logger.Debug("Created sprint details", "sprint", *sprint)
	return sprint, nil
}

func (r *sprintRepository) FindByID(ctx context.Context, id int) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "FindByID",
		"sprint_id", id,
	)
	logger.Debug("Starting find sprint by ID process")

	s := r.q.Sprint
	sprint, err := s.WithContext(ctx).
		Where(s.ID.Eq(id)).
		Preload(s.Tasks).
		Preload(s.Project).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Sprint not found")
			return nil, structs.ErrSprintNotExist
		}
		logger.Error("Failed to find sprint by ID due to database error", "error", err)
		return nil, structs.ErrDatabaseFail
	}

	logger.Info("Successfully found sprint by ID")
	logger.Debug("Successfully retrieved sprint with associations", "sprintID", sprint.ID, "taskCount", len(sprint.Tasks), "projectName", sprint.Project.Name) // Example using preloaded data
	return sprint, nil
}

func (r *sprintRepository) Find(ctx context.Context, filter *dto.SprintFilter) ([]*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Find",
	)
	logger.Debug("Starting find sprints process", "filter", filter)

	s := r.q.Sprint
	sprintQuery := s.WithContext(ctx)

	if filter.ID != nil {
		logger.Debug("Applying filter: ID", "sprint_id", *filter.ID)
		sprintQuery = sprintQuery.Where(s.ID.Eq(*filter.ID))
	}
	if filter.Name != nil && *filter.Name != "" {
		namePattern := fmt.Sprintf("%%%s%%", *filter.Name)
		logger.Debug("Applying filter: Name", "sprint_name_pattern", namePattern)
		sprintQuery = sprintQuery.Where(s.Name.Like(namePattern))
	}
	if filter.ProjectID != nil {
		logger.Debug("Applying filter: ProjectID", "project_id", *filter.ProjectID)
		sprintQuery = sprintQuery.Where(s.ProjectID.Eq(*filter.ProjectID))
	}
	if filter.StartDateAfter != nil {
		logger.Debug("Applying filter: StartDateAfter", "start_date", filter.StartDateAfter)
		sprintQuery = sprintQuery.Where(s.StartDate.Gte(*filter.StartDateAfter))
	}
	if filter.EndDateBefore != nil {
		logger.Debug("Applying filter: EndDateBefore", "end_date", filter.EndDateBefore)
		sprintQuery = sprintQuery.Where(s.EndDate.Lte(*filter.EndDateBefore))
	}

	sprints, err := sprintQuery.Find()
	if err != nil {
		logger.Error("Error finding sprints", "error", err)
		return nil, fmt.Errorf("database error retrieving sprints: %w", err)
	}

	logger.Info("Successfully found sprints", "count", len(sprints))
	logger.Debug("Found sprints details", "sprints", sprints)

	return sprints, nil
}

func (r *sprintRepository) Update(ctx context.Context, id int, updateMap map[string]any) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Update",
		"sprint_id", id,
	)
	logger.Debug("Starting update sprint process", "update_data", updateMap)

	s := r.q.Sprint
	resultInfo, err := s.WithContext(ctx).Where(s.ID.Eq(id)).Updates(updateMap)

	if err != nil {
		logger.Error("Failed to update sprint", "error", err)
		return fmt.Errorf("failed to update sprint %d: %w", id, err)
	}

	if resultInfo.RowsAffected == 0 {
		logger.Warn("Update executed but no sprint found with the given ID or data was the same")
		return structs.ErrSprintNotExist
	}

	logger.Info("Successfully updated sprint", "rows_affected", resultInfo.RowsAffected)
	return nil
}

func (r *sprintRepository) Delete(ctx context.Context, id int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintRepository",
		"method", "Delete",
		"sprint_id", id,
	)
	logger.Debug("Starting delete sprint process")

	s := r.q.Sprint
	resultInfo, err := s.WithContext(ctx).Where(s.ID.Eq(id)).Delete()

	if err != nil {
		logger.Error("Failed to delete sprint due to database error", "error", err)
		return structs.ErrDatabaseFail
	}

	if resultInfo.RowsAffected == 0 {
		logger.Warn("Delete executed but no sprint found with the given ID")
		return structs.ErrSprintNotExist
	}

	logger.Info("Successfully deleted sprint", "rows_affected", resultInfo.RowsAffected)
	return nil
}