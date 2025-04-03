package handler

import (
	"errors"
	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	taskService service.TaskService
	cfg         config.DateTimeConfig
}

func NewTaskHandler(taskService service.TaskService, cfg config.DateTimeConfig) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		cfg:         cfg,
	}
}

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
		logger.Error("Can not parsing input", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", nil))
	}
	errs := utils.ValidateStruct(*input)
	if errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}

	logger.Debug("Validation successful", "input", *input)
	userClaims, _ := c.Locals("user_claims").(*structs.Claims)
	task := input.MapToTask()

	task, err := h.taskService.CreateTask(ctx, userClaims.UserID, task.SprintID, task)
	if err != nil {
		if errors.Is(err, structs.ErrSprintNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("sprint not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("user not authorize", err.Error()))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse("internal error", nil))
		}
	}

	output := dto.MapToTaskResponse(task)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("successfully created sprint", output))
}

func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "GetTask",
	)

	id, err := verifyIdParamInt(c, logger, "taskId")
	if err != nil {
		logger.Error("Invalid task id")
		return err
	}

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	task, err := h.taskService.FindByID(ctx, userClaims.UserID, id)
	if err != nil {
		if errors.Is(err, structs.ErrTaskNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("task not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("internal server error", nil))
		} else {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("forbiden", err.Error()))
		}
	}

	output := dto.MapToTaskResponse(task)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("successfully found task", output))
}

func (h *TaskHandler) FindTasksByProjectID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "FindTasksByProjectID",
	)

	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		logger.Error("Invalid project id")
		return err
	}

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	tasks, err := h.taskService.FindTasksByProjectID(ctx, userClaims.UserID, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(createErrorResponse("can not find project", err.Error()))
		}
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(createErrorResponse("forbiden", err.Error()))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse("internal server error", nil))
	}

	output := dto.MapToSliceOfTaskResponse(tasks)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusAccepted).JSON(createSliceSuccessResponseGeneric("successfully found tasks", output))
}

func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "TaskHandler",
		"handler", "DeleteTask",
	)
	taskID, err := verifyIdParamInt(c, logger, "taskId")
	if err != nil {
		logger.Error("Invalid task id")
		return err
	}

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	if err := h.taskService.DeleteTask(ctx, userClaims.UserID, taskID); err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		} else {
			if errors.Is(err, structs.ErrTaskNotExist) {
				return c.Status(fiber.StatusNotFound).JSON(createErrorResponse("Task does not exist", err.Error()))
			} else {
				return c.Status(fiber.StatusForbidden).JSON(createErrorResponse("Authorize failed", err.Error()))
			}
		}
	}
	logger.Info("Delete task successfully", "task_id", taskID)

	return c.Status(fiber.StatusAccepted).JSON(createErrorResponse("Delete sprint successfully", nil))
}
