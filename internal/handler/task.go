package handler

import (
	"errors"
	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	taskService service.TaskService
	cfg         config.DateTimeConfig
}

// NewTaskHandler creates a new TaskHandler instance
func NewTaskHandler(taskService service.TaskService, cfg config.DateTimeConfig) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		cfg:         cfg,
	}
}

// CreateTask creates a new task
// @Summary Create a new task
// @Description Creates a new task with the provided details for a specific sprint
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body dto.CreateTaskRequest true "Task creation request"
// @Success 201 {object} dto.TaskSuccessResponse "Task created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Sprint not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "CreateTask",
	)

	logger.Debug("Parsing input...")
	input := &dto.CreateTaskRequest{}
	if err := c.BodyParser(input); err != nil {
		logger.Error("Cannot parse input", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "errors", errs)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}

	logger.Debug("Validation successful", "input", *input)
	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	task := input.MapToTask()
	task, err := h.taskService.CreateTask(ctx, userClaims.UserID, task.SprintID, task)
	if err != nil {
		if errors.Is(err, structs.ErrSprintNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Sprint not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("User not authorized", err.Error()))
		}
		logger.Error("Failed to create task", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	output := dto.MapToTaskResponse(task)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusCreated).JSON(createSuccessResponse("Task created successfully", output))
}

// GetTask retrieves a task by ID
// @Summary Get a task by ID
// @Description Retrieves details of a specific task
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param taskId path int true "Task ID"
// @Success 202 {object} dto.TaskSuccessResponse "Task found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid task ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Task not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /tasks/{taskId} [get]
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "GetTask",
	)

	id, err := verifyIdParamInt(c, logger, "taskId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	task, err := h.taskService.FindByID(ctx, userClaims.UserID, id)
	if err != nil {
		if errors.Is(err, structs.ErrTaskNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Task not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		}
		return c.Status(fiber.StatusForbidden).JSON(
			createErrorResponse("Forbidden", err.Error()))
	}

	output := dto.MapToTaskResponse(task)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse("Task found successfully", output))
}

// FindTasksByUserID retrieves tasks assigned to a user
// @Summary Get tasks by user ID
// @Description Retrieves all tasks assigned to a specific user
// @Tags Tasks
// @Produce json
// @Param userId path int true "User ID"
// @Success 202 {object} dto.TaskSliceSuccessResponse "Tasks found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid user ID"
// @Failure 404 {object} dto.ErrorResponse "Not found - User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{userId}/tasks [get]
func (h *TaskHandler) FindTasksByUserID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "FindTasksByUserID",
	)

	id, err := verifyIdParamInt(c, logger, "userId")
	if err != nil {
		return err
	}

	tasks, err := h.taskService.FindTasksByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("User not found", err.Error()))
		}
		logger.Error("Failed to find tasks", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	output := dto.MapToSliceOfTaskResponse(tasks)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSliceSuccessResponseGeneric("Tasks found successfully", output))
}

// FindTasksByProjectID retrieves tasks for a project
// @Summary Get tasks by project ID
// @Description Retrieves all tasks associated with a specific project
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Success 202 {object} dto.TaskSliceSuccessResponse "Tasks found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid project ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Project not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects/{projectId}/tasks [get]
func (h *TaskHandler) FindTasksByProjectID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "FindTasksByProjectID",
	)

	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	tasks, err := h.taskService.FindTasksByProjectID(ctx, userClaims.UserID, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Project not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("Forbidden", err.Error()))
		}
		logger.Error("Failed to find tasks", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	output := dto.MapToSliceOfTaskResponse(tasks)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSliceSuccessResponseGeneric("Tasks found successfully", output))
}

// UpdateTask updates an existing task
// @Summary Update a task
// @Description Updates the details of an existing task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param taskId path int true "Task ID"
// @Param task body dto.UpdateTaskRequest true "Task update request"
// @Success 202 {object} dto.TaskSuccessResponse "Task updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or task ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Task not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /tasks/{taskId} [put]
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "UpdateTask",
	)

	id, err := verifyIdParamInt(c, logger, "taskId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	input := &dto.UpdateTaskRequest{}
	if err := c.BodyParser(input); err != nil {
		logger.Error("Cannot parse input", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "errors", errs)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}

	logger.Debug("Validation successful", "input", *input)
	updatedTask, err := h.taskService.UpdateTask(ctx, userClaims.UserID, id, input)
	if err != nil {
		if errors.Is(err, structs.ErrTaskNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Task not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("Forbidden", err.Error()))
		}
		logger.Error("Failed to update task", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	output := dto.MapToTaskResponse(updatedTask)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse("Task updated successfully", output))
}

// AssignTaskToUser assigns a task to a user
// @Summary Assign task to user
// @Description Assigns a specific task to a user
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param taskId path int true "Task ID"
// @Param userId path int true "User ID"
// @Success 202 {object} dto.GenericSuccessResponse "Task assigned successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid IDs or user not in project"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Task or user not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /tasks/{taskId}/user/{userId} [post]
func (h *TaskHandler) AssignTaskToUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "AssignTaskToUser",
	)

	taskID, err := verifyIdParamInt(c, logger, "taskId")
	if err != nil {
		return err
	}

	userID, err := verifyIdParamInt(c, logger, "userId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	if err := h.taskService.AssignTaskToUser(ctx, userID, userClaims.UserID, taskID); err != nil {
		if errors.Is(err, structs.ErrTaskNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Task not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("User not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("Forbidden", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotPartProject) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("User is not part of the project", err.Error()))
		}
		logger.Error("Failed to assign task", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	logger.Info("Task assigned to user successfully", "task_id", taskID, "user_id", userID)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse[any]("Task assigned successfully", nil))
}

// FindTasks retrieves tasks based on filters
// @Summary Find tasks with filters
// @Description Retrieves tasks based on optional query parameters (id, title, status, priority, due_date_before)
// @Tags Tasks
// @Produce json
// @Param id query int false "Task ID"
// @Param title query string false "Task title"
// @Param status query string false "Task status" Enums(OPEN, IN_PROGRESS, DONE)
// @Param priority query string false "Task priority" Enums(LOW, MEDIUM, HIGH)
// @Param due_date_before query string false "Due date before (format: YYYY-MM-DD)"
// @Success 202 {object} dto.TaskSliceSuccessResponse "Tasks found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /tasks [get]
func (h *TaskHandler) FindTasks(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "FindTasks",
	)

	filter := &dto.TaskFilter{}
	var parseErrors []string

	if idStr := c.Query("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Error("Invalid id parameter", "id", idStr)
			parseErrors = append(parseErrors, "Invalid id parameter")
		} else {
			filter.ID = &id
		}
	}
	if title := c.Query("title"); title != "" {
		filter.Title = &title
		logger.Debug("Valid title parameter", "title", title)
	}
	if status := c.Query("status"); status != "" {
		taskStatus := models.TaskStatus(status)
		filter.Status = &taskStatus
		logger.Debug("Valid status parameter", "status", status)
	}
	if priority := c.Query("priority"); priority != "" {
		taskPriority := models.TaskPriority(priority)
		filter.Priority = &taskPriority
		logger.Debug("Valid priority parameter", "priority", priority)
	}
	if dueDateBefore := c.Query("due_date_before"); dueDateBefore != "" {
		dueDate, err := time.Parse(h.cfg.Format, dueDateBefore)
		if err != nil {
			logger.Error("Invalid due_date_before parameter", "due_date_before", dueDateBefore)
			parseErrors = append(parseErrors, "Invalid due_date_before parameter")
		} else {
			filter.DueDateBefore = &dueDate
		}
	}

	if len(parseErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", parseErrors))
	}

	logger.Debug("Validation successful", "filter", *filter)
	tasks, err := h.taskService.FindTasks(ctx, filter)
	if err != nil {
		logger.Error("Service error finding tasks", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to retrieve tasks", nil))
	}

	output := dto.MapToSliceOfTaskResponse(tasks)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(
		createSliceSuccessResponseGeneric("Tasks found successfully", output))
}

// DeleteTask deletes a task by ID
// @Summary Delete a task
// @Description Deletes a specific task
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param taskId path int true "Task ID"
// @Success 202 {object} dto.GenericSuccessResponse "Task deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid task ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Task not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /tasks/{taskId} [delete]
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "DeleteTask",
	)

	taskID, err := verifyIdParamInt(c, logger, "taskId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	if err := h.taskService.DeleteTask(ctx, userClaims.UserID, taskID); err != nil {
		if errors.Is(err, structs.ErrTaskNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Task not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		}
		return c.Status(fiber.StatusForbidden).JSON(
			createErrorResponse("Forbidden", err.Error()))
	}

	logger.Info("Task deleted successfully", "task_id", taskID)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse[any]("Task deleted successfully", nil))
}