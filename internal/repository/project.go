package repository

import (
	"context"
	"fmt"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/query"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) (*models.Project, error)
	Find(ctx context.Context, filter dto.ProjectFilter) ([]*models.Project, error)
	FindByID(ctx context.Context, id int) (*models.Project, error)
	Update(ctx context.Context, id int, updateMap map[string]any) error
	Delete(ctx context.Context, id int) error
}

type projectRepository struct {
	db  *gorm.DB
	cfg config.DateTimeConfig
	q   *query.Query
	*GenericRepository[*models.Project, int]
}

func NewProjectRepository(db *gorm.DB, cfg config.DateTimeConfig) ProjectRepository {
	genericRepo := NewGenericRepository[*models.Project, int](
		db,
		"Project",
		structs.ErrProjectNotExist,
	)

	return &projectRepository{
		db:  db,
		cfg: cfg,
		q:   query.Use(db),
		GenericRepository: genericRepo,
	}
}

// func (r *projectRepository) Create(ctx context.Context, project *models.Project) (*models.Project, error) {
// 	baseLogger := utils.LoggerFromContext(ctx)
// 	logger := baseLogger.With(
// 		"component", "ProjectRepository",
// 		"method", "Create",
// 	)
// 	logger.Debug("Starting create project process", "project_name", project.Name)

// 	err := r.q.Project.WithContext(ctx).Create(project)
// 	if err != nil {
// 		logger.Error("Failed to create project", "error", err)
// 		return nil, structs.ErrDataViolateConstraint
// 	}

// 	logger.Info("Successfully created project", "project_id", project.ID)
// 	return project, nil
// }

func (r *projectRepository) Find(ctx context.Context, filter dto.ProjectFilter) ([]*models.Project, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectRepository",
		"method", "Find",
	)
	logger.Debug("Starting find projects process", "filter", filter)

	p := r.q.Project
	projectQuery := p.WithContext(ctx)

	if filter.ID != nil {
		logger.Debug("Applying filter: ID", "project_id", *filter.ID)
		projectQuery = projectQuery.Where(p.ID.Eq(*filter.ID))
	}
	if filter.Name != nil && *filter.Name != "" {
		namePattern := fmt.Sprintf("%%%s%%", *filter.Name)
		logger.Debug("Applying filter: Name", "project_name_pattern", namePattern)
		projectQuery = projectQuery.Where(p.Name.Like(namePattern))
	}
	if filter.Status != nil {
		logger.Debug("Applying filter: Status", "project_status", *filter.Status)
		projectQuery = projectQuery.Where(p.Status.Eq(string(*filter.Status)))
	}
	if filter.ManagerID != nil {
		logger.Debug("Applying filter: ManagerID", "manager_id", *filter.ManagerID)
		projectQuery = projectQuery.Where(p.ManagerID.Eq(*filter.ManagerID))
	}
	if filter.StartDateAfter != nil {
		logger.Debug("Applying filter: StartDateAfter", "start_date", filter.StartDateAfter)
		projectQuery = projectQuery.Where(p.StartDate.Gte(*filter.StartDateAfter))
	}
	if filter.EndDateBefore != nil {
		logger.Debug("Applying filter: EndDateBefore", "end_date", filter.EndDateBefore)
		projectQuery = projectQuery.Where(p.EndDate.Lte(*filter.EndDateBefore))
	}

	projects, err := projectQuery.Find()
	if err != nil {
		logger.Error("Error finding projects", "error", err)
		return nil, fmt.Errorf("database error retrieving projects: %w", err)
	}

	logger.Info("Successfully found projects", "count", len(projects))
	logger.Debug("Found projects details", "projects", projects) // could be large

	return projects, nil
}

// func (r *projectRepository) FindByID(ctx context.Context, id int) (*models.Project, error) {
// 	baseLogger := utils.LoggerFromContext(ctx)
// 	logger := baseLogger.With(
// 		"component", "ProjectRepository",
// 		"method", "FindByID",
// 		"project_id", id,
// 	)
// 	logger.Debug("Starting find project by ID process")

// 	p := r.q.Project
// 	project, err := p.WithContext(ctx).Where(p.ID.Eq(id)).First()

// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			logger.Warn("Project not found")
// 			return nil, structs.ErrProjectNotExist
// 		}
// 		logger.Error("Failed to find project by ID due to database error", "error", err)
// 		return nil, fmt.Errorf("database error finding project %d: %w", id, err)
// 	}

// 	logger.Info("Successfully found project by ID")
// 	return project, nil
// }

// func (r *projectRepository) Update(ctx context.Context, id int, updateMap map[string]any) error {
// 	baseLogger := utils.LoggerFromContext(ctx)
// 	logger := baseLogger.With(
// 		"component", "ProjectRepository",
// 		"method", "Update",
// 		"project_id", id,
// 	)
// 	logger.Debug("Starting update project process", "update_data", updateMap)

// 	p := r.q.Project
// 	resultInfo, err := p.WithContext(ctx).Where(p.ID.Eq(id)).Updates(updateMap)

// 	if err != nil {
// 		logger.Error("Failed to update project", "error", err)
// 		return fmt.Errorf("failed to update project %d: %w", id, err)
// 	}

// 	if resultInfo.RowsAffected == 0 {
// 		logger.Warn("Update executed but no project found with the given ID or data was the same")
// 		return structs.ErrProjectNotExist
// 	}

// 	logger.Info("Successfully updated project", "rows_affected", resultInfo.RowsAffected)
// 	return nil
// }

// func (r *projectRepository) Delete(ctx context.Context, id int) error {
// 	baseLogger := utils.LoggerFromContext(ctx)
// 	logger := baseLogger.With(
// 		"component", "ProjectRepository",
// 		"method", "Delete",
// 		"project_id", id,
// 	)
// 	logger.Debug("Starting delete project process")

// 	p := r.q.Project
// 	resultInfo, err := p.WithContext(ctx).Where(p.ID.Eq(id)).Delete()

// 	if err != nil {
// 		logger.Error("Failed to delete project due to database error", "error", err)
// 		return structs.ErrDatabaseFail
// 	}

// 	if resultInfo.RowsAffected == 0 {
// 		logger.Warn("Delete executed but no project found with the given ID")
// 		return structs.ErrProjectNotExist
// 	}

// 	logger.Info("Successfully deleted project", "rows_affected", resultInfo.RowsAffected)
// 	return nil
// }
