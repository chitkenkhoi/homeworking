package repository

import (
	"context"
	"errors"
	"fmt"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/query"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) (*models.Task, error)
	FindByID(ctx context.Context, id int) (*models.Task, error)
	FindTasksByProjectID(ctx context.Context, projectID int) ([]*models.Task, error)
	Delete(ctx context.Context, id int) error
}

type taskRepository struct {
	db  *gorm.DB
	cfg config.DateTimeConfig
	q   *query.Query
}

func NewTaskRepository(db *gorm.DB, cfg config.DateTimeConfig) TaskRepository {
	return &taskRepository{
		db:  db,
		cfg: cfg,
		q:   query.Use(db),
	}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) (*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "Create",
	)
	logger.Debug("Starting create task process", "task_tittle", task.Title)

	err := r.q.Task.WithContext(ctx).Create(task)
	if err != nil {
		logger.Error("Failed to create task", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}

	logger.Info("Successfully created task", "task_id", task.ID)
	logger.Debug("Created task details", "task", *task)
	return task, nil
}

func (r *taskRepository) FindByID(ctx context.Context, id int) (*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "FindByID",
		"task_id", id,
	)
	logger.Debug("Starting find task by ID process")

	s := r.q.Task
	task, err := s.WithContext(ctx).
		Where(s.ID.Eq(id)).
		Preload(s.Assignee).
		Preload(s.Project).
		Preload(s.Sprint).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Task not found")
			return nil, structs.ErrTaskNotExist
		}
		logger.Error("Failed to find task by ID due to database error", "error", err)
		return nil, structs.ErrDatabaseFail
	}

	logger.Info("Successfully found task by ID")
	logger.Debug("Successfully retrieved task with associations", "task", task.ID, "sprintName", task.Sprint.Name, "projectName", task.Project.Name) // Example using preloaded data
	return task, nil
}

func (r *taskRepository) FindTasksByProjectID(ctx context.Context, projectID int) ([]*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "FindTasksByProjectID",
		"project_id", projectID,
	)
	logger.Debug("Starting find tasks by project ID process")
	t := r.q.Task 

	taskQuery := t.WithContext(ctx).
		Where(t.ProjectID.Eq(projectID)).
		Preload(t.Assignee)

	tasks, err := taskQuery.Find()
	if err != nil {
		logger.Error("Failed to find tasks by project ID due to database error", "error", err)
		return nil, fmt.Errorf("database error finding tasks for project %d: %w", projectID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully found tasks for project", "count", len(tasks))
	return tasks, nil
}

func (r *taskRepository) Delete(ctx context.Context, id int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "Delete",
		"task_id", id,
	)
	logger.Debug("Starting delete task process")

	s := r.q.Task
	resultInfo, err := s.WithContext(ctx).Where(s.ID.Eq(id)).Delete()

	if err != nil {
		logger.Error("Failed to delete task due to database error", "error", err)
		return structs.ErrDatabaseFail
	}

	if resultInfo.RowsAffected == 0 {
		logger.Warn("Delete executed but no sprint found with the given ID")
		return structs.ErrTaskNotExist
	}

	logger.Info("Successfully deleted task", "rows_affected", resultInfo.RowsAffected)
	return nil
}
