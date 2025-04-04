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

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) (*models.Task, error)
	AssignTaskToUser(ctx context.Context, userID, taskID int) error
	Find(ctx context.Context, filter *dto.TaskFilter) ([]*models.Task, error)
	FindByID(ctx context.Context, id int) (*models.Task, error)
	Update(ctx context.Context, id int, updateMap map[string]any) error
	FindTasksByProjectID(ctx context.Context, projectID int) ([]*models.Task, error)
	FindTaskByUserID(ctx context.Context, userID int) ([]*models.Task, error)
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

func (r *taskRepository) FindTaskByUserID(ctx context.Context, userID int) ([]*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "FindTaskByUserID",
		"user_id", userID,
	)
	logger.Debug("Starting find tasks by user ID process")
	t := r.q.Task
	taskQuery := t.WithContext(ctx).
		Where(t.AssigneeID.Eq(userID)).
		Preload(t.Assignee)
	tasks, err := taskQuery.Find()
	if err != nil {
		logger.Error("Failed to find tasks by user ID due to database error", "error", err)
		return nil, fmt.Errorf("database error finding tasks for user %d: %w", userID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully found tasks for user", "count", len(tasks))
	return tasks, nil
}

func (r *taskRepository) Find(ctx context.Context, filter *dto.TaskFilter) ([]*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "Find",
	)
	logger.Debug("Starting find tasks process", "filter", filter)

	t := r.q.Task
	taskQuery := t.WithContext(ctx)

	if filter.ID != nil {
		logger.Debug("Applying filter: ID", "task_id", *filter.ID)
		taskQuery = taskQuery.Where(t.ID.Eq(*filter.ID))
	}
	if filter.Title != nil && *filter.Title != "" {
		titlePattern := fmt.Sprintf("%%%s%%", *filter.Title)
		logger.Debug("Applying filter: Title", "task_title_pattern", titlePattern)
		taskQuery = taskQuery.Where(t.Title.Like(titlePattern))
	}
	if filter.Status != nil {
		logger.Debug("Applying filter: Status", "task_status", *filter.Status)
		taskQuery = taskQuery.Where(t.Status.Eq(string(*filter.Status)))
	}
	if filter.Priority != nil {
		logger.Debug("Applying filter: Priority", "task_priority", *filter.Priority)
		taskQuery = taskQuery.Where(t.Priority.Eq(string(*filter.Priority)))
	}
	if filter.DueDateBefore != nil {
		logger.Debug("Applying filter: DueDateBefore", "due_date", filter.DueDateBefore)
		taskQuery = taskQuery.Where(t.DueDate.Lte(*filter.DueDateBefore))
	}

	tasks,err := taskQuery.Find()
	if err != nil {
		logger.Error("Error finding tasks", "error", err)
		return nil, fmt.Errorf("database error retrieving tasks: %w", err)
	}
	logger.Info("Successfully found tasks", "count", len(tasks))
	logger.Debug("Found tasks details", "tasks", tasks)
	return tasks, nil
}

func (r *taskRepository) AssignTaskToUser(ctx context.Context, userID, taskID int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "AssignTaskToUser",
		"user_id", userID,
		"task_id", taskID,
	)
	logger.Debug("Starting assign task to user process")
	s := r.q.Task
	task, err := s.WithContext(ctx).Where(s.ID.Eq(taskID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Task not found")
			return structs.ErrTaskNotExist
		}
		logger.Error("Failed to find task by ID due to database error", "error", err)
		return structs.ErrDatabaseFail
	}
	task.AssigneeID = &userID
	resultInfo, err := s.WithContext(ctx).Where(s.ID.Eq(taskID)).Updates(task)
	if err != nil {
		logger.Error("Failed to assign task to user due to database error", "error", err)
		return structs.ErrDatabaseFail
	}

	logger.Info("Successfully updated sprint", "rows_affected", resultInfo.RowsAffected)
	return nil
}

func (r *taskRepository) Update(ctx context.Context, id int, updateMap map[string]any) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskRepository",
		"method", "Update",
		"task_id", id,
	)
	logger.Debug("Starting update task process", "update_data", updateMap)

	s := r.q.Task
	resultInfo, err := s.WithContext(ctx).Where(s.ID.Eq(id)).Updates(updateMap)

	if err != nil {
		logger.Error("Failed to update task", "error", err)
		return fmt.Errorf("failed to update task %d: %w", id, err)
	}

	if resultInfo.RowsAffected == 0 {
		logger.Warn("Update executed but no task found with the given ID or data was the same")
		return structs.ErrTaskNotExist
	}

	logger.Info("Successfully updated task", "rows_affected", resultInfo.RowsAffected)
	return nil
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
