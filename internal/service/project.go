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

type ProjectService interface {
	CreateProject(ctx context.Context, project *models.Project) (*models.Project, error)
	ListProjects(ctx context.Context, filter dto.ProjectFilter) ([]*models.Project, error)
	FindByID(ctx context.Context, id int) (*models.Project, error)
	AddTeamMembers(ctx context.Context, userID, projectID int, userIDsToAdd []int) (int, error)
	UpdateProject(ctx context.Context, userID, projectId int, data *dto.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, userID, projectID int) error
	GetAndVerifyProjectManager(ctx context.Context, userID, projectID int) (*models.Project, error)
}

type projectService struct {
	projectRepository repository.ProjectRepository
	userService       UserService
}

func NewProjectService(projectRepository repository.ProjectRepository, userService UserService) ProjectService {
	return &projectService{
		projectRepository: projectRepository,
		userService: userService,
	}
}

func (s *projectService) GetAndVerifyProjectManager(ctx context.Context, userID, projectID int) (*models.Project, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectService",
		"method", "getAndVerifyProjectManager",
		"project_id", projectID,
		"requestor_id", userID,
	)

	logger.Debug("Fetching project by ID")
	project, err := s.projectRepository.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			logger.Warn(structs.MsgProjectNotExist)
			return nil, structs.ErrProjectNotExist
		}
		logger.Error(structs.MsgInternalDatabaseErrFetchingProject, "error", err)
		return nil, structs.ErrDatabaseFail
	}

	logger.Debug(structs.MsgVerifyingProjectManager)
	if project.ManagerID != userID {
		logger.Warn(structs.MsgAuthorizationFailure,
			"manager_id", project.ManagerID)
		return nil, structs.ErrUserNotManageProject
	}

	logger.Debug(structs.MsgProjectAuthorized)
	return project, nil
}

func (s *projectService) CreateProject(ctx context.Context, project *models.Project) (*models.Project, error) {
	return s.projectRepository.Create(ctx, project)
}

func (s *projectService) ListProjects(ctx context.Context, filter dto.ProjectFilter) ([]*models.Project, error) {
	projects, err := s.projectRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

func (s *projectService) FindByID(ctx context.Context, id int) (*models.Project, error) {
	project, err := s.projectRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			slog.Error("Project does not exist", "id", id)
			return nil, err
		}
		return nil, structs.ErrDatabaseFail
	}
	slog.Info("Find project with id", "id", id, "data", project)
	return project, nil
}

func (s *projectService) AddTeamMembers(ctx context.Context, userID, projectID int, userIDsToAdd []int) (int, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectService",
		"method", "AddTeamMembers",
		"project_id", projectID,
		"requestor_id", userID,
		"users_to_add", userIDsToAdd,
	)

	logger.Debug("Starting team member addition process")

	project, err := s.projectRepository.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			logger.Error("Project not found", "error", err)
			return 0, err
		}
		logger.Error("Failed to retrieve project", "error", err)
		return 0, err
	}

	logger.Debug("Project found", "project_manager_id", project.ManagerID)

	if managerID := project.ManagerID; managerID != userID {
		logger.Error("Authorization failure",
			"error", structs.ErrUserNotManageProject,
			"requestor_id", userID,
			"manager_id", managerID,
		)
		return 0, structs.ErrUserNotManageProject
	}

	logger.Debug("Validating users for assignment")
	validUserIDs, validationErr := s.userService.FindValidTeamMembersForAssignment(ctx, userIDsToAdd)

	if validationErr != nil {
		logger.Debug("User validation produced errors",
			"valid_users_count", len(validUserIDs),
		)

		if len(validUserIDs) == 0 {
			return 0, fmt.Errorf("%w: underlying reason: %v", structs.ErrNoValidUserStatus, validationErr)
		}

		logger.Error("Some users failed validation",
			"valid_users_count", len(validUserIDs),
			"requested_users_count", len(userIDsToAdd),
		)
	} else {
		logger.Debug("All users validated successfully", "valid_users_count", len(validUserIDs))
	}

	logger.Debug("Assigning users to project", "user_count", len(validUserIDs))

	err = s.userService.AssignUsersToProject(ctx, projectID, validUserIDs)
	if err != nil {
		logger.Error("Failed to assign users to project", "error", err)
		return 0, structs.ErrDatabaseFail
	}

	logger.Info("Successfully added users to project",
		"project_id", projectID,
		"user_count", len(validUserIDs),
	)

	return len(validUserIDs), validationErr
}

func (s *projectService) UpdateProject(ctx context.Context, userID, projectID int, data *dto.UpdateProjectRequest) (*models.Project, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectService",
		"method", "UpdateProject",
		"project_id", projectID,
		"user_id", userID,
	)

	logger.Debug("Starting to update project")

	project, err := s.GetAndVerifyProjectManager(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return nil, fmt.Errorf("cannot update project: %w with id %d", err, projectID)
		}
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("user %d cannot update project %d: %w", userID, projectID, err)
		}
		logger.Error("Failed initial project retrieval or authorization", "error", err)
		return nil, err
	}

	logger.Debug("Starting to validate end date")

	if data.EndDate != nil {
		endDateValue := *data.EndDate
		if !endDateValue.After(project.StartDate) {
			logger.Error("End date must be after start date",
				"end_date", endDateValue,
				"start_date", project.StartDate)
			return nil, structs.ErrEndDateBeforeStartDate
		}
	}

	logger.Debug("End date is valid",
		"end_date", *data.EndDate,
		"start_date", project.StartDate)

	updateMap := make(map[string]any)
	if data.Name != nil {
		updateMap["name"] = *data.Name
	}
	if data.Description != nil {
		updateMap["description"] = *data.Description
	}
	if data.EndDate != nil {
		updateMap["end_date"] = data.EndDate
	}
	if data.Status != nil {
		updateMap["status"] = *data.Status
	}
	if len(updateMap) == 0 {
		logger.Info("No fields to update, returning current project")
		return project, nil
	}

	logger.Debug("Attempting project update operation", "input", updateMap)

	if err := s.projectRepository.Update(ctx, projectID, updateMap); err != nil {
		logger.Error("Failed to update project in repository", "error", err)
		return nil, structs.ErrDatabaseFail
	}

	logger.Info("Succesfully updated")

	updatedProject, _ := s.projectRepository.FindByID(ctx, projectID)
	return updatedProject, nil
}

func (s *projectService) DeleteProject(ctx context.Context, userID, projectID int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "ProjectService",
		"method", "DeleteProject",
		"project_id", projectID,
		"requestor_id", userID,
	)

	logger.Debug("Starting project deletion process")

	_, err := s.GetAndVerifyProjectManager(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return fmt.Errorf("cannot delete project: %w with id %d", err, projectID)
		}
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return fmt.Errorf("user %d cannot delete project %d: %w", userID, projectID, err)
		}
		logger.Error("Failed initial project retrieval or authorization for deletion", "error", err)
		return err
	}

	logger.Debug("Authorization successful, attempting project deletion")
	if err := s.projectRepository.Delete(ctx, projectID); err != nil {
		logger.Error("Failed to delete project in repository", "error", err)
		return fmt.Errorf("repository delete failed for project %d: %w", projectID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully deleted project")
	return nil
}
