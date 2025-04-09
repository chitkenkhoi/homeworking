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

// SprintHandler handles sprint-related HTTP requests
type SprintHandler struct {
	sprintService service.SprintService
	cfg           config.DateTimeConfig
}

// NewSprintHandler creates a new SprintHandler instance
func NewSprintHandler(sprintService service.SprintService, cfg config.DateTimeConfig) *SprintHandler {
	return &SprintHandler{
		sprintService: sprintService,
		cfg:           cfg,
	}
}

// CreateSprint creates a new sprint
// @Summary Create a new sprint
// @Description Creates a new sprint for a specific project
// @Tags Sprints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sprint body dto.CreateSprintRequest true "Sprint creation request"
// @Success 201 {object} dto.SprintSuccessResponse "Sprint created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or dates"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Project not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /sprints [post]
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

	sprint := input.MapToSprint()
	sprint, err := h.sprintService.CreateSprint(ctx, userClaims.UserID, sprint.ProjectID, sprint)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Project not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("User not authorized", err.Error()))
		} else if errors.Is(err, structs.ErrSprintDateInvalid) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("Sprint dates are not valid", err.Error()))
		}
		logger.Error("Failed to create sprint", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	output := dto.MapToSprintResponse(sprint)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusCreated).JSON(createSuccessResponse("Sprint created successfully", output))
}

// GetSprint retrieves a sprint by ID
// @Summary Get a sprint by ID
// @Description Retrieves details of a specific sprint
// @Tags Sprints
// @Produce json
// @Security BearerAuth
// @Param sprintId path int true "Sprint ID"
// @Success 200 {object} dto.SprintSuccessResponse "Sprint found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid sprint ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Sprint not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /sprints/{sprintId} [get]
func (h *SprintHandler) GetSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "GetSprint",
	)

	id, err := verifyIdParamInt(c, logger, "sprintId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	sprint, err := h.sprintService.FindByID(ctx, userClaims.UserID, id)
	if err != nil {
		if errors.Is(err, structs.ErrSprintNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Sprint not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		}
		return c.Status(fiber.StatusForbidden).JSON(
			createErrorResponse("Forbidden", err.Error()))
	}

	output := dto.MapToSprintResponse(sprint)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse("Sprint found successfully", output))
}

// FindSprints retrieves sprints based on filters
// @Summary Find sprints with filters
// @Description Retrieves sprints based on optional query parameters (id, name, projectid, startdate, enddate)
// @Tags Sprints
// @Produce json
// @Param id query int false "Sprint ID"
// @Param name query string false "Sprint name"
// @Param projectid query int false "Project ID"
// @Param startdate query string false "Start date after (format: YYYY-MM-DD)"
// @Param enddate query string false "End date before (format: YYYY-MM-DD)"
// @Success 200 {object} dto.SprintSliceSuccessResponse "Sprints found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /sprints [get]
func (h *SprintHandler) FindSprints(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "FindSprints",
	)

	filter := &dto.SprintFilter{}
	var parseErrors []string

	if idStr := c.Query("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Error("Invalid id parameter", "id", idStr)
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'id' parameter: %s", idStr))
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
			logger.Error("Invalid projectid parameter", "projectid", projectIdStr)
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'projectid' parameter: %s", projectIdStr))
		} else {
			logger.Debug("Valid projectid parameter", "projectid", projectId)
			filter.ProjectID = &projectId
		}
	}

	if startDateStr := c.Query("startdate"); startDateStr != "" {
		startDate, err := time.Parse(h.cfg.Format, startDateStr)
		if err != nil {
			logger.Error("Invalid startdate parameter", "startdate", startDateStr)
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'startdate' format (use YYYY-MM-DD): %s", startDateStr))
		} else {
			logger.Debug("Valid startdate parameter", "startdate", startDate)
			filter.StartDateAfter = &startDate
		}
	}

	if endDateStr := c.Query("enddate"); endDateStr != "" {
		endDate, err := time.Parse(h.cfg.Format, endDateStr)
		if err != nil {
			logger.Error("Invalid enddate parameter", "enddate", endDateStr)
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'enddate' format (use YYYY-MM-DD): %s", endDateStr))
		} else {
			logger.Debug("Valid enddate parameter", "enddate", endDate)
			filter.EndDateBefore = &endDate
		}
	}

	if len(parseErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Invalid query parameters", parseErrors))
	}

	sprints, err := h.sprintService.FindSprints(ctx, filter)
	if err != nil {
		logger.Error("Service error finding sprints", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to retrieve sprints", nil))
	}

	logger.Debug("Successfully retrieved sprints", "sprints", sprints)
	outputs := dto.MapToSprintResponseSlice(sprints)
	logger.Debug("Response is prepared", "response", outputs)
	return c.Status(fiber.StatusOK).JSON(
		createSliceSuccessResponseGeneric("Sprints found successfully", outputs))
}

// UpdateSprint updates an existing sprint
// @Summary Update a sprint
// @Description Updates the details of an existing sprint
// @Tags Sprints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sprintId path int true "Sprint ID"
// @Param sprint body dto.UpdateSprintRequest true "Sprint update request"
// @Success 200 {object} dto.SprintSuccessResponse "Sprint updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or sprint ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /sprints/{sprintId} [put]
func (h *SprintHandler) UpdateSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "UpdateSprint",
	)

	sprintID, err := verifyIdParamInt(c, logger, "sprintId")
	if err != nil {
		return err
	}

	input := &dto.UpdateSprintRequest{}
	if err = c.BodyParser(input); err != nil {
		logger.Error("Cannot parse JSON", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "errors", errs)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Validation failed", errs))
	}

	logger.Debug("Validation successful", "input", *input)

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	updatedSprint, err := h.sprintService.UpdateSprint(ctx, userClaims.UserID, sprintID, input)
	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			logger.Error("Database failure", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal database failure", nil))
		} else if errors.Is(err, structs.ErrSprintNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Sprint not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("Forbidden", err.Error()))
		}
		logger.Error("Failed to update sprint", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation error", err.Error()))
	}

	output := dto.MapToSprintResponse(updatedSprint)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse("Sprint updated successfully", output))
}

// DeleteSprint deletes a sprint by ID
// @Summary Delete a sprint
// @Description Deletes a specific sprint
// @Tags Sprints
// @Produce json
// @Security BearerAuth
// @Param sprintId path int true "Sprint ID"
// @Success 200 {object} dto.GenericSuccessResponse "Sprint deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid sprint ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Sprint not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /sprints/{sprintId} [delete]
func (h *SprintHandler) DeleteSprint(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintHandler",
		"handler", "DeleteSprint",
	)

	sprintID, err := verifyIdParamInt(c, logger, "sprintId")
	if err != nil {
		return err
	}

	userClaims, ok := c.Locals("user_claims").(*structs.Claims)
	if !ok {
		logger.Error("Failed to retrieve user claims")
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	if err := h.sprintService.DeleteSprint(ctx, userClaims.UserID, sprintID); err != nil {
		if errors.Is(err, structs.ErrSprintNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Sprint not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			logger.Error("Database failure", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		}
		logger.Error("Failed to delete sprint", "error", err.Error())
		return c.Status(fiber.StatusForbidden).JSON(
			createErrorResponse("Forbidden", err.Error()))
	}

	logger.Info("Sprint deleted successfully", "sprint_id", sprintID)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse[any]("Sprint deleted successfully", nil))
}