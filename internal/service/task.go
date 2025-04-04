package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/repository"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"
)

type TaskService interface {
	CreateTask(ctx context.Context, userID, sprintID int, task *models.Task) (*models.Task, error)
	AssignTaskToUser(ctx context.Context, userID, reqID, taskID int) error
	FindByID(ctx context.Context, userID, taskID int) (*models.Task, error)
	UpdateTask(ctx context.Context, userID, taskID int, data *dto.UpdateTaskRequest) (*models.Task, error)
	FindTasksByUserID(ctx context.Context, userID int) ([]*models.Task, error)
	FindTasksByProjectID(ctx context.Context, userID, projectID int) ([]*models.Task, error)
	FindTasks(ctx context.Context, filter *dto.TaskFilter) ([]*models.Task, error)
	DeleteTask(ctx context.Context, userID, taskID int) error
}

type taskService struct {
	taskRepository repository.TaskRepository
	projectService ProjectService
	sprintService  SprintService
	userService    UserService
}

func NewTaskService(taskRepository repository.TaskRepository, projectService ProjectService, sprintService SprintService, userService UserService) TaskService {
	return &taskService{
		taskRepository: taskRepository,
		projectService: projectService,
		sprintService:  sprintService,
	}
}

func (s *taskService) GetAndVerifyProjectManagerForTask(ctx context.Context, baseLogger *slog.Logger, userID, taskID int, isCommand bool) (*models.Task, error) {
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
			logger.Error("Database error during task fetch", "original_error", err)
			return nil, structs.ErrDatabaseFail
		}
	}

	logger.Debug("Task retrieval success", "task_id", task.ID, "project_id", task.Project.ID, "manager_id", task.Project.ManagerID, "assignee_id", task.AssigneeID)

	logger.Debug("Starting authorize user")
	isManager := task.Project.ManagerID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID

	if isCommand {
		logger.Debug("Command operation: Checking for manager privileges")
		if !isManager {
			logger.Warn("Authorization failed: User is not project manager for command operation",
				"required_manager_id", task.Project.ManagerID)
			return nil, structs.ErrUserNotManageProject
		}
		logger.Debug("Authorization success: User is project manager")
	} else {
		logger.Debug("Query operation: Checking for manager or assignee privileges")
		if !isManager && !isAssignee {
			logger.Warn("Authorization failed: User is neither project manager nor task assignee for query operation",
				"required_manager_id", task.Project.ManagerID,
				"task_assignee_id", task.AssigneeID)
			return nil, structs.ErrUserNotAuthorizedForTask
		}
		if isManager {
			logger.Debug("Authorization success: User is project manager")
		} else {
			logger.Debug("Authorization success: User is task assignee")
		}
	}

	logger.Info("Task retrieval and user authorization successful")
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
	task.SprintID = sprintID

	task, err = s.taskRepository.Create(ctx, task)
	if err != nil {
		logger.Error("Repository failed to create task", "erorr", err)
		return nil, structs.ErrDatabaseFail
	}

	task.Sprint = sprint
	return task, nil
}

func (s *taskService) AssignTaskToUser(ctx context.Context, userID, reqID, taskID int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "AssignTaskToUser",
		"task_id", taskID,
		"requestor_id", reqID,
		"user_id", userID,
	)

	logger.Debug("Starting assign task to user process")
	task, err := s.GetAndVerifyProjectManagerForTask(ctx, logger, reqID, taskID, true)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) || errors.Is(err, structs.ErrUserNotAuthorizedForTask) {
			return fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return fmt.Errorf("cannot fetch task: %w with task id: %d", err, taskID)
		}
	}
	logger.Debug("Task retrieval success", "task_id", task.ID, "project_id", task.Project.ID, "manager_id", task.Project.ManagerID, "assignee_id", task.AssigneeID)
	logger.Info("Verify user id to assign task")
	user, err := s.userService.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			logger.Error("User not found", "error", err)
		} else {
			logger.Error("Failed to retrieve user", "error", err)
		}
		return fmt.Errorf("cannot assign task: %w with user id %d", err, userID)
	}

	if user.CurrentProjectID == nil || *user.CurrentProjectID != task.ProjectID {
		logger.Warn("User is not part of the project", "user_id", userID, "project_id", task.ProjectID)
		return structs.ErrUserNotPartProject
	}

	if err := s.taskRepository.AssignTaskToUser(ctx, userID, task.ID); err != nil {
		logger.Error("Failed to assign task to user in repository", "error", err)
		return fmt.Errorf("repository failed to assign task to user %d: %w", userID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully assigned task to user")
	return nil
}

func (s *taskService) UpdateTask(ctx context.Context, userID, taskID int, data *dto.UpdateTaskRequest) (*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "UpdateTask",
		"task_id", taskID,
		"requestor_id", userID,
	)

	logger.Debug("Starting task update process")
	task, err := s.GetAndVerifyProjectManagerForTask(ctx, logger, userID, taskID, true)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) || errors.Is(err, structs.ErrUserNotAuthorizedForTask) {
			return nil, fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return nil, fmt.Errorf("cannot fetch task: %w with task id: %d", err, taskID)
		}
	}
	updateMap := make(map[string]any)
	if data.Title != nil {
		updateMap["title"] = *data.Title
	}
	if data.Description != nil {
		updateMap["description"] = *data.Description
	}
	if data.Status != nil {
		updateMap["status"] = *data.Status
	}
	if data.Priority != nil {
		updateMap["priority"] = *data.Priority
	}
	if data.DueDate != nil {
		updateMap["due_date"] = data.DueDate
	}

	if len(updateMap) == 0 {
		logger.Info("No fields to update, returning current sprint")
		return task, nil
	}

	logger.Debug("Attempting task update operation", "input", updateMap)

	if err := s.taskRepository.Update(ctx, taskID, updateMap); err != nil {
		logger.Error("Failed to update task in repository", "error", err)
		return nil, fmt.Errorf("repository failed to update task %d: %w", task.ID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully updated task")

	updatedTask, err := s.taskRepository.FindByID(ctx, taskID)
	if err != nil {
		return task, nil
	}
	return updatedTask, nil
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

	task, err := s.GetAndVerifyProjectManagerForTask(ctx, logger, userID, taskID, false)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) || errors.Is(err, structs.ErrUserNotAuthorizedForTask) {
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

func (s *taskService) FindTasksByUserID(ctx context.Context, userID int) ([]*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "FindTasksByUserID",
		"user_id", userID,
	)
	

	logger.Info("Starting task retreival process")
	tasks, err := s.taskRepository.FindTaskByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) FindTasks(ctx context.Context, filter *dto.TaskFilter) ([]*models.Task, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskService",
		"method", "FindTasks",
	)

	tasks, err := s.taskRepository.Find(ctx, filter)
	if err != nil {
		logger.Error("Failed to find tasks by filter", "error", err)
		return nil, fmt.Errorf("database error finding tasks with filter: %w", err)
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
	_, err := s.GetAndVerifyProjectManagerForTask(ctx, logger, userID, taskID, true)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) || errors.Is(err, structs.ErrUserNotAuthorizedForTask) {
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
