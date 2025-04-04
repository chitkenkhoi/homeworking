package handler

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type SprintHandler struct {
	sprintService service.SprintService
	cfg           config.DateTimeConfig
}

func NewSprintHandler(sprintService service.SprintService, cfg config.DateTimeConfig) *SprintHandler {
	return &SprintHandler{
		sprintService: sprintService,
		cfg:           cfg,
	}
}

func (h *SprintHandler) CreateSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "CreateSprint",
	)

	logger.Debug("Parsing input...")
	input := &dto.CreateSprintRequest{}
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
	sprint := input.MapToSprint()

	sprint, err := h.sprintService.CreateSprint(ctx, userClaims.UserID, sprint.ProjectID, sprint)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("project not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("user not authorize", err.Error()))
		} else if errors.Is(err, structs.ErrSprintDateInvalid) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("sprint's date are not valid", err.Error()))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse("internal error", nil))
		}
	}

	output := dto.MapToSprintResponse(sprint)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("successfully created sprint", output))
}

func (h *SprintHandler) GetSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "GetSprint",
	)

	id, err := verifyIdParamInt(c, logger, "sprintId")
	if err != nil {
		logger.Error("Invalid sprint id")
		return err
	}

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	sprint, err := h.sprintService.FindByID(ctx, userClaims.UserID, id)
	if err != nil {
		if errors.Is(err, structs.ErrSprintNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("sprint not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("internal server error", nil))
		} else {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("forbiden", err.Error()))
		}
	}

	output := dto.MapToSprintResponse(sprint)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("successfully found sprint", output))
}

func (h *SprintHandler) FindSprints(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "FindSprints",
	)

	filter := &dto.SprintFilter{}
	var err error
	var parseErrors []string
	if idStr := c.Query("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Error("Invalid id paramter", "id", idStr)
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'id' parameter: %s", idStr))
		} else {
			logger.Debug("Valid id parameter", "id", id)
			filter.ID = &id
		}
	}

	if nameStr := c.Query("name"); nameStr != "" {
		filter.Name = &nameStr
		logger.Debug("Valid name parameter", "name", nameStr)
	}

	if projectIdStr := c.Query("projectid"); projectIdStr != "" {
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			logger.Error("Invalid projectid paramter", "projectid", projectIdStr)
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'projectid' parameter: %s", projectIdStr))
		} else {
			logger.Debug("Valid projectid parameter", "projectid", projectId)
			filter.ProjectID = &projectId
		}
	}

	if startDateStr := c.Query("startdate"); startDateStr != "" {
		startDate, err := time.Parse(h.cfg.Format, startDateStr)
		if err != nil {
			logger.Error("Invalid startdate paramter", "startdate", startDateStr)
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'startdate' format (use YYYY-MM-DD): %s", startDateStr))
		} else {
			logger.Debug("Valid startdate parameter", "startdate", startDate)
			filter.StartDateAfter = &startDate
		}
	}

	if endDateStr := c.Query("enddate"); endDateStr != "" {
		endDate, err := time.Parse(h.cfg.Format, endDateStr)
		if err != nil {
			logger.Error("Invalid enddate paramter", "enddate", endDateStr)
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'enddate' format (use YYYY-MM-DD): %s", endDateStr))
		} else {
			logger.Debug("Valid enddate parameter", "enddate", endDate)
			filter.EndDateBefore = &endDate
		}
	}

	if len(parseErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("invalid format input", parseErrors))
	}

	sprints, err := h.sprintService.FindSprints(ctx, filter)
	if err != nil {
		logger.Error("Service error finding sprints", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to retrieve sprints", nil))
	}

	logger.Debug("Successfully retrieve sprints", "srpints", sprints)

	outputs := dto.MapToSprintResponseSlice(sprints)

	logger.Debug("Response is prepared", "response", outputs)
	return c.Status(fiber.StatusOK).JSON(
		createSliceSuccessResponseGeneric("success", outputs))
}

func (h *SprintHandler) UpdateSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "UpdateSprint",
	)

	sprintID, err := verifyIdParamInt(c, logger, "sprintId")
	if err != nil {
		logger.Error("Invalid sprint id")
		return err
	}

	input := &dto.UpdateSprintRequest{}
	if err = c.BodyParser(input); err != nil {
		logger.Error("Cannot parse JSON", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "errors", errs)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Validation failed", nil))
	}

	logger.Debug("Validation finish successfully for input", "input", *input)

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	updatedSprint, err := h.sprintService.UpdateSprint(ctx, userClaims.UserID, sprintID, input)
	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse("Internal database fail", nil))
		}
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Validation error", err.Error()))
	}
	// //optimize memory here
	//input = nil
	output := dto.MapToSprintResponse(updatedSprint)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("Sprint has been updated", output))
}

func (h *SprintHandler) DeleteSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "DeleteSprint",
	)
	sprintID, err := verifyIdParamInt(c, logger, "sprintId")
	if err != nil {
		logger.Error("Invalid sprint id")
		return err
	}

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	if err := h.sprintService.DeleteSprint(ctx, userClaims.UserID, sprintID); err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		} else {
			if errors.Is(err, structs.ErrSprintNotExist) {
				return c.Status(fiber.StatusNotFound).JSON(createErrorResponse("Sprint does not exist", err.Error()))
			} else {
				return c.Status(fiber.StatusForbidden).JSON(createErrorResponse("Authorize failed", err.Error()))
			}
		}
	}
	logger.Info("Delete sprint successfully", "sprint_id", sprintID)

	return c.Status(fiber.StatusAccepted).JSON(createErrorResponse("Delete sprint successfully", nil))
}
