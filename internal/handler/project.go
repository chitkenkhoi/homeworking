package handler

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

func (h *ProjectHandler) CreateProjectHandler(c *fiber.Ctx) error {
	input := &dto.CreateProjectRequest{}
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", err.Error()))
	}
	errors := utils.ValidateStruct(*input) // Pass the struct value
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errors))
	}
	log.Printf("Validation successful for input: %+v\n", *input)

	userClaims, _ := c.Locals("user_claims").(*structs.Claims)
	project := input.MapToProject(userClaims.UserID)

	if project, err := h.projectService.CreateProject(project); err != nil {
		log.Printf("Failed to create project: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to create project", err.Error()))
	} else {
		output := &dto.ProjectResponse{}
		output.MapToDto(project)
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

	dateFormat := "2006-01-02"

	if startDateStr := c.Query("startdate"); startDateStr != "" {
		startDate, err := time.Parse(dateFormat, startDateStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid 'startdate' format (use YYYY-MM-DD): %s", startDateStr))
		} else {
			// Optional: Set to start of the day if needed
			// startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
			filter.StartDateAfter = &startDate
		}
	}

	if endDateStr := c.Query("enddate"); endDateStr != "" {
		endDate, err := time.Parse(dateFormat, endDateStr)
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

	projects, err := h.projectService.ListProjects(filter)
	if err != nil {
		log.Printf("Service error listing projects: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to retrieve projects", nil))
	}
	output := make([]dto.ProjectResponse, len(projects))

	for i, p := range projects {
		output[i].MapToDto(&p)
	}
	return c.Status(fiber.StatusOK).JSON(
		createSliceSuccessResponseGeneric("success", output))
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	return nil
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	return nil
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	return nil
}
