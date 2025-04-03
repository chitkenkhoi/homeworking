package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/repository"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"
)

type TaskService interface {
	CreateTask(ctx context.Context, userID, sprintID int, task *models.Task) (*models.Task, error)
	FindByID(ctx context.Context, userID, taskID int) (*models.Task, error)
	FindTasksByProjectID(ctx context.Context, userID, projectID int) ([]*models.Task, error)
	DeleteTask(ctx context.Context, userID, taskID int) error
}

type taskService struct {
	taskRepository repository.TaskRepository
	projectService ProjectService
	sprintService  SprintService
}

func NewTaskService(taskRepository repository.TaskRepository, projectService ProjectService, sprintService SprintService) TaskService {
	return &taskService{
		taskRepository: taskRepository,
		projectService: projectService,
		sprintService:  sprintService,
	}
}

func (s *taskService) GetAndVerifyProjectManagerForTask(ctx context.Context, baseLogger *slog.Logger, userID, taskID int) (*models.Task, error) {
	logger := baseLogger.With(
		"method", "GetAndVerifyProjectManagerForTask",
	)

	logger.Debug("Starting retreive task")

	task, err := s.taskRepository.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Can not fetch task", "error", err)
		if errors.Is(err, structs.ErrTaskNotExist) {
			return nil, structs.ErrTaskNotExist
		} else {
			return nil, structs.ErrDatabaseFail
		}
	}

	logger.Debug("Task retreival success", "task", *task)

	logger.Debug("Starting authorize user")

	if task.Project.ManagerID != userID {
		logger.Error("MsgAuthorizationFailure",
			"manager_id", task.Project.ManagerID)
		return nil, structs.ErrUserNotManageProject
	}

	logger.Info("Task retreival success and user is authorized")
	return task, nil
}

func (s *taskService) CreateTask(ctx context.Context, userID, sprintID int, task *models.Task) (*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "CreateTask",
		"sprint_id", sprintID,
		"requestor_id", userID,
	)

	logger.Debug("Start verify project manager")

	sprint, err := s.sprintService.GetAndVerifyProjectManagerForSprint(ctx, logger, userID, sprintID)
	if err != nil {
		if errors.Is(err, structs.ErrSprintNotExist) {
			return nil, fmt.Errorf("cannot create task: %w with sprint id %d", err, sprintID)
		}
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("user %d cannot create task in sprint %d: %w", userID, sprintID, err)
		}
		logger.Error("Failed initial project retrieval or authorization", "error", err)
		return nil, err
	}

	task.ProjectID = sprint.ProjectID
	task.SprintID = sprintID //ensure consistency

	task, err = s.taskRepository.Create(ctx, task)
	if err != nil {
		logger.Error("Repository failed to create task", "erorr", err)
		return nil, structs.ErrDatabaseFail
	}

	task.Sprint = sprint
	return task, nil
}

func (s *taskService) FindByID(ctx context.Context, userID, taskID int) (*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintService",
		"method", "FindByID",
		"task_id", taskID,
		"requestor_id", userID,
	)

	logger.Debug("Starting retreive sprint")

	task, err := s.GetAndVerifyProjectManagerForTask(ctx, logger, userID, taskID)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return nil, fmt.Errorf("cannot fetch task: %w with task id: %d", err, taskID)
		}
	}

	return task, nil
}

func (s *taskService) FindTasksByProjectID(ctx context.Context, userID, projectID int) ([]*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "FindTasksByProjectID",
		"project_id", projectID,
		"requestor_id", userID,
	)

	logger.Info("Starting task retreival process")
	_, err := s.projectService.GetAndVerifyProjectManager(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return nil, fmt.Errorf("cannot find project: %w with id %d", err, projectID)
		}
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("user %d cannot query project %d: %w", userID, projectID, err)
		}
		logger.Error("Failed initial project retrieval or authorization", "error", err)
		return nil, err
	}

	tasks, err := s.taskRepository.FindTasksByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) DeleteTask(ctx context.Context, userID, taskID int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "DeleteTask",
		"task_id", taskID,
		"requestor_id", userID,
	)

	logger.Info("Starting task deletion process")
	_, err := s.GetAndVerifyProjectManagerForTask(ctx, logger, userID, taskID)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return fmt.Errorf("cannot fetch task: %w with task id: %d", err, taskID)
		}
	}

	logger.Info("Authorization successful, attempting task deletion")

	if err := s.taskRepository.Delete(ctx, taskID); err != nil {
		logger.Error("Failed to delete task in repository", "error", err)
		return fmt.Errorf("repository delete failed for task %d: %w", taskID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully deleted task")
	return nil
}
