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

type ProjectHandler struct {
	projectService service.ProjectService
	cfg            config.DateTimeConfig
}

func NewProjectHandler(projectService service.ProjectService, cfg config.DateTimeConfig) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		cfg:            cfg,
	}
}

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
	project := input.MapToProject(userClaims.UserID)
	// //optimize memory here
	//input = nil

	if project, err := h.projectService.CreateProject(ctx, project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to create project", nil))
	} else {
		output := dto.MapToProjectDto(project)
		logger.Debug("Response is prepared", "response", output)
		return c.Status(fiber.StatusCreated).JSON(
			createSuccessResponse("project is created", output))
	}
}

func (h *ProjectHandler) ListProjectsHanlder(c *fiber.Ctx) error {
	filter := dto.ProjectFilter{}
	var err error
	var parseErrors []string
	if idStr := c.Query("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'id' parameter: %s", idStr))
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
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'status' parameter: %s", statusStr))
		} else {
			filter.Status = &status
		}
	}

	if managerIdStr := c.Query("managerid"); managerIdStr != "" {
		managerId, err := strconv.Atoi(managerIdStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'managerid' parameter: %s", managerIdStr))
		} else {
			filter.ManagerID = &managerId
		}
	}

	if startDateStr := c.Query("startdate"); startDateStr != "" {
		startDate, err := time.Parse(h.cfg.Format, startDateStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'startdate' format (use YYYY-MM-DD): %s", startDateStr))
		} else {
			filter.StartDateAfter = &startDate
		}
	}

	if endDateStr := c.Query("enddate"); endDateStr != "" {
		endDate, err := time.Parse(h.cfg.Format, endDateStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'enddate' format (use YYYY-MM-DD): %s", endDateStr))
		} else {
			filter.EndDateBefore = &endDate
		}
	}

	if len(parseErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("invalid format input", parseErrors))
	}

	ctx := c.UserContext()
	projects, err := h.projectService.ListProjects(ctx, filter)
	if err != nil {
		log.Printf("Service error listing projects: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to retrieve projects", nil))
	}

	outputs := dto.MapToProjectDtoSlice(projects)
	return c.Status(fiber.StatusOK).JSON(
		createSliceSuccessResponseGeneric("success", outputs))
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "GetProject",
	)

	id, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		logger.Error("Invalid project id")
		return err
	}

	project, err := h.projectService.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("project is not found",
					fmt.Errorf("project with id %v does not exist", id)))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("internal server error", nil))
		}

	} else {
		output := dto.MapToProjectDto(project)

		logger.Debug("Data response is availabe", "data", output)

		return c.Status(fiber.StatusFound).JSON(
			createSuccessResponse("found project", output))
	}
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "UpdateProject",
	)

	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		logger.Error("Invalid project id")
		return err
	}

	input := &dto.UpdateProjectRequest{}
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

	updatedProject, err := h.projectService.UpdateProject(ctx, userClaims.UserID, projectID, input)
	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse("Internal database fail", nil))
		}
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Validation error", err.Error()))
	}
	// //optimize memory here
	//input = nil
	output := dto.MapToProjectDto(updatedProject)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("Project has been updated", output))
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "DeleteProject",
	)
	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		logger.Error("Invalid project id")
		return err
	}

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	if err := h.projectService.DeleteProject(ctx, userClaims.UserID, projectID); err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", nil))
		} else {
			if errors.Is(err, structs.ErrProjectNotExist) {
				return c.Status(fiber.StatusNotFound).JSON(createErrorResponse("Project does not exist", err.Error()))
			} else {
				return c.Status(fiber.StatusForbidden).JSON(createErrorResponse("Authorize failed", err.Error()))
			}
		}
	}
	logger.Info("Delete project successfully", "project_id", projectID)

	return c.Status(fiber.StatusAccepted).JSON(createErrorResponse("Delete project successfully", nil))
}

func (h *ProjectHandler) AddTeamMembers(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "ProjectHandler",
		"handler", "AddTeamMembersHandler",
	)

	projectID, err := verifyIdParamInt(c, logger, "projectId")
	if err != nil {
		logger.Error("Invalid project id")
		return err
	}

	input := &dto.AddTeamMembersRequest{}
	if err = c.BodyParser(input); err != nil {
		logger.Error("Cannot parse JSON", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "errors", errs)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Validation failed", nil))
	}

	logger.Debug("Validation finish successfully for input", "input", input)

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)

	if count, err := h.projectService.AddTeamMembers(ctx, userClaims.UserID, projectID, input.UserIDs); err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) ||
			errors.Is(err, structs.ErrUserNotManageProject) ||
			errors.Is(err, structs.ErrNoValidUserStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("Could not add team members due to validation errors.", err.Error()))
		} else if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse("Internal server error", nil))
		} else {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"message":       "Some of the users are not valid to be assigned to the project",
				"detail":        err.Error(),
				"updated_count": count,
			})
		}
	} else {
		return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse("All users are added to the project", count))
	}

}
