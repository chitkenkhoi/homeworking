package handler

import (
	"errors"
	"fmt"
	"log"

	"strconv"
	"time"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	projectService service.ProjectService
	cfg            config.DateTimeConfig
}

// NewProjectHandler creates a new ProjectHandler instance
func NewProjectHandler(projectService service.ProjectService, cfg config.DateTimeConfig) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		cfg:            cfg,
	}
}

// CreateProjectHandler creates a new project
// @Summary Create a new project
// @Description Creates a new project with the provided details
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project body dto.CreateProjectRequest true "Project creation request"
// @Success 201 {object} dto.ProjectSuccessResponse "Project created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects [post]
func (h *ProjectHandler) CreateProjectHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "CreateProjectHandler",
	)

	logger.Debug("Parsing input...")
	input := &dto.CreateProjectRequest{}
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

	project := input.MapToProject(userClaims.UserID)
	project, err := h.projectService.CreateProject(ctx, project)
	if err != nil {
		logger.Error("Failed to create project", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to create project", err.Error()))
	}

	output := dto.MapToProjectDto(project)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusCreated).JSON(
		createSuccessResponse("Project created successfully", output))
}

// ListProjectsHanlder lists projects based on filters
// @Summary List projects with filters
// @Description Retrieves projects based on optional query parameters (id, name, status, managerid, startdate, enddate)
// @Tags Projects
// @Produce json
// @Param id query int false "Project ID"
// @Param name query string false "Project name"
// @Param status query string false "Project status" Enums(ACTIVE, INACTIVE, COMPLETED)
// @Param managerid query int false "Manager ID"
// @Param startdate query string false "Start date after (format: YYYY-MM-DD)"
// @Param enddate query string false "End date before (format: YYYY-MM-DD)"
// @Success 200 {object} dto.ProjectSliceSuccessResponse "Projects found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects [get]
func (h *ProjectHandler) ListProjectsHanlder(c *fiber.Ctx) error {
	filter := dto.ProjectFilter{}
	var parseErrors []string

	if idStr := c.Query("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'id' parameter: %s", idStr))
		} else {
			filter.ID = &id
		}
	}

	if nameStr := c.Query("name"); nameStr != "" {
		filter.Name = &nameStr
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.ProjectStatus(statusStr)
		if !status.IsValid() {
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'status' parameter: %s", statusStr))
		} else {
			filter.Status = &status
		}
	}

	if managerIdStr := c.Query("managerid"); managerIdStr != "" {
		managerId, err := strconv.Atoi(managerIdStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'managerid' parameter: %s", managerIdStr))
		} else {
			filter.ManagerID = &managerId
		}
	}

	if startDateStr := c.Query("startdate"); startDateStr != "" {
		startDate, err := time.Parse(h.cfg.Format, startDateStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'startdate' format (use YYYY-MM-DD): %s", startDateStr))
		} else {
			filter.StartDateAfter = &startDate
		}
	}

	if endDateStr := c.Query("enddate"); endDateStr != "" {
		endDate, err := time.Parse(h.cfg.Format, endDateStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("Invalid 'enddate' format (use YYYY-MM-DD): %s", endDateStr))
		} else {
			filter.EndDateBefore = &endDate
		}
	}

	if len(parseErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Invalid query parameters", parseErrors))
	}

	ctx := c.UserContext()
	projects, err := h.projectService.ListProjects(ctx, filter)
	if err != nil {
		log.Printf("Service error listing projects: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to retrieve projects", err.Error()))
	}

	outputs := dto.MapToProjectDtoSlice(projects)
	return c.Status(fiber.StatusOK).JSON(
		createSliceSuccessResponseGeneric("Projects found successfully", outputs))
}

// GetProject retrieves a project by ID
// @Summary Get a project by ID
// @Description Retrieves details of a specific project
// @Tags Projects
// @Produce json
// @Param projectId path int true "Project ID"
// @Success 200 {object} dto.ProjectSuccessResponse "Project found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid project ID"
// @Failure 404 {object} dto.ErrorResponse "Not found - Project not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects/{projectId} [get]
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "GetProject",
	)

	id, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		return err
	}

	project, err := h.projectService.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Project not found",
					fmt.Errorf("Project with ID %v does not exist", id)))
		}
		logger.Error("Failed to find project", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal server error", nil))
	}

	output := dto.MapToProjectDto(project)
	logger.Debug("Response is prepared", "data", output)
	return c.Status(fiber.StatusOK).JSON(
		createSuccessResponse("Project found successfully", output))
}

// UpdateProject updates an existing project
// @Summary Update a project
// @Description Updates the details of an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Param project body dto.UpdateProjectRequest true "Project update request"
// @Success 200 {object} dto.ProjectSuccessResponse "Project updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or project ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects/{projectId} [put]
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "UpdateProject",
	)

	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		return err
	}

	input := &dto.UpdateProjectRequest{}
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

	updatedProject, err := h.projectService.UpdateProject(ctx, userClaims.UserID, projectID, input)
	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			logger.Error("Database failure", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal database failure", nil))
		} else if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Project not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("Forbidden", err.Error()))
		}
		logger.Error("Failed to update project", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation error", err.Error()))
	}

	output := dto.MapToProjectDto(updatedProject)
	logger.Debug("Response is prepared", "response", output)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse("Project updated successfully", output))
}

// DeleteProject deletes a project by ID
// @Summary Delete a project
// @Description Deletes a specific project
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Success 200 {object} dto.GenericSuccessResponse "Project deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid project ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 404 {object} dto.ErrorResponse "Not found - Project not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects/{projectId} [delete]
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "DeleteProject",
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

	if err := h.projectService.DeleteProject(ctx, userClaims.UserID, projectID); err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Project not found", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			logger.Error("Database failure", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		}
		logger.Error("Failed to delete project", "error", err.Error())
		return c.Status(fiber.StatusForbidden).JSON(
			createErrorResponse("Forbidden", err.Error()))
	}

	logger.Info("Project deleted successfully", "project_id", projectID)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse[any]("Project deleted successfully", nil))
}

// AddTeamMembers adds team members to a project
// @Summary Add team members to a project
// @Description Adds one or more users to a project as team members
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Param members body dto.AddTeamMembersRequest true "List of user IDs to add"
// @Success 200 {object} dto.IntSuccessResponse "Team members added successfully"
// @Success 207 {object} dto.AddTeamMembersPartialSuccessResponse "Some team members added successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or project ID"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User not authorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /projects/{projectId}/members [post]
func (h *ProjectHandler) AddTeamMembers(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "AddTeamMembers",
	)

	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		return err
	}

	input := &dto.AddTeamMembersRequest{}
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

	count, err := h.projectService.AddTeamMembers(ctx, userClaims.UserID, projectID, input.UserIDs)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("Project not found", err.Error()))
		} else if errors.Is(err, structs.ErrUserNotManageProject) {
			return c.Status(fiber.StatusForbidden).JSON(
				createErrorResponse("Forbidden", err.Error()))
		} else if errors.Is(err, structs.ErrNoValidUserStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("No valid users to add", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			logger.Error("Database failure", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		}
		logger.Warn("Some users could not be added", "error", err.Error(), "count", count)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":       "Some users could not be added to the project",
			"details":       err.Error(),
			"updated_count": count,
		})
	}

	logger.Info("All team members added successfully", "project_id", projectID, "count", count)
	return c.Status(fiber.StatusOK).JSON(createSuccessResponse("All team members added successfully", count))
}