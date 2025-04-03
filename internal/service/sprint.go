package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/repository"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"
)

type SprintService interface {
	CreateSprint(ctx context.Context, userID, projectID int, sprint *models.Sprint) (*models.Sprint, error)
	FindByID(ctx context.Context, userID, sprintID int) (*models.Sprint, error)
	ListSprints(ctx context.Context, filter dto.SprintFilter) ([]*models.Sprint, error)
	GetAndVerifyProjectManagerForSprint(ctx context.Context, baseLogger *slog.Logger, userID, sprintID int) (*models.Sprint, error)
	UpdateSprint(ctx context.Context, userID, sprintID int, data *dto.UpdateSprintRequest) (*models.Sprint, error)
	DeleteSprint(ctx context.Context, userID, sprintID int) error
}

type sprintService struct {
	projectRepository repository.ProjectRepository
	sprintRepository  repository.SprintRepository
	projectService    ProjectService
	cfg               config.DateTimeConfig
}

func NewSprintService(projectRepository repository.ProjectRepository,
	sprintRepository repository.SprintRepository,
	projectService ProjectService,
	cfg config.DateTimeConfig) SprintService {
	return &sprintService{
		projectRepository: projectRepository,
		sprintRepository:  sprintRepository,
		projectService:    projectService,
		cfg:               cfg,
	}
}

func (s *sprintService) GetAndVerifyProjectManagerForSprint(ctx context.Context, baseLogger *slog.Logger, userID, sprintID int) (*models.Sprint, error) {
	logger := baseLogger.With(
		"method", "GetAndVerifyProjectManagerForSprint",
	)

	logger.Debug("Starting retreive sprint")

	sprint, err := s.sprintRepository.FindByID(ctx, sprintID)
	if err != nil {
		logger.Error("Can not fetch sprint", "error", err.Error())
		if errors.Is(err, structs.ErrSprintNotExist) {
			return nil, structs.ErrSprintNotExist
		} else {
			return nil, structs.ErrDatabaseFail
		}
	}

	logger.Debug("Sprint retreival success", "sprint", *sprint)

	logger.Debug("Starting authorize user")

	if sprint.Project.ManagerID != userID {
		logger.Error("MsgAuthorizationFailure",
			"manager_id", sprint.Project.ManagerID)
		return nil, structs.ErrUserNotManageProject
	}

	logger.Info("Sprint retreival success and user is authorized")
	return sprint, nil
}

func (s *sprintService) CreateSprint(ctx context.Context, userID, projectID int, sprint *models.Sprint) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintService",
		"method", "CreateSprint",
		"project_id", projectID,
		"requestor_id", userID,
	)

	logger.Debug("Start verify project manager")

	project, err := s.projectService.GetAndVerifyProjectManager(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, structs.ErrProjectNotExist) {
			return nil, fmt.Errorf("cannot create sprint: %w with project id %d", err, projectID)
		}
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("user %d cannot create sprint in project %d: %w", userID, projectID, err)
		}
		logger.Error("Failed initial project retrieval or authorization", "error", err)
		return nil, err
	}

	logger.Debug("Start validate sprint date")

	if sprint.StartDate.Before(project.StartDate) {
		logger.Error("Sprint's start date must not after project's start date",
			"project_start", project.StartDate, "sprint_start", sprint.StartDate)
		return nil, fmt.Errorf("cannot create sprint: %w with start date %s", structs.ErrSprintDateInvalid, sprint.StartDate.Format(s.cfg.Format))
	} else {
		if project.EndDate != nil && sprint.StartDate.After(*project.EndDate) {
			logger.Error("Sprint's end date must not before project's end date",
				"project_end", project.EndDate, "sprint_end", sprint.EndDate)
			return nil, fmt.Errorf("cannot create sprint: %w with end date %s", structs.ErrSprintDateInvalid, sprint.EndDate.Format(s.cfg.Format))
		}
	}

	logger.Debug("Sprint date are valid")

	sprint, err = s.sprintRepository.Create(ctx, sprint)
	if err != nil {
		logger.Error("Repository failed to create sprint", "error", err)
		return nil, structs.ErrDatabaseFail
	}

	return sprint, nil
}

func (s *sprintService) ListSprints(ctx context.Context, filter dto.SprintFilter) ([]*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintService",
		"method", "ListSprints",
	)

	sprints, err := s.sprintRepository.Find(ctx, filter)
	if err != nil {
		logger.Error("Failed to list projects", "error", err)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return sprints, nil
}

func (s *sprintService) FindByID(ctx context.Context, userID, sprintID int) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintService",
		"method", "FindByID",
		"sprint_id", sprintID,
		"requestor_id", userID,
	)

	logger.Debug("Starting retreive sprint")

	sprint, err := s.GetAndVerifyProjectManagerForSprint(ctx, logger, userID, sprintID)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return nil, fmt.Errorf("cannot fetch sprint: %w with sprint id: %d", err, sprintID)
		}
	}

	return sprint, nil
}

func (s *sprintService) UpdateSprint(ctx context.Context, userID, sprintID int, data *dto.UpdateSprintRequest) (*models.Sprint, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintService",
		"method", "UpdateSprint",
		"sprint_id", sprintID,
		"requestor_id", userID,
	)
	logger.Debug("Starting sprint update process")
	sprint, err := s.GetAndVerifyProjectManagerForSprint(ctx, logger, userID, sprintID)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return nil, fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return nil, fmt.Errorf("cannot fetch sprint: %w with sprint id: %d", err, sprintID)
		}
	}

	logger.Debug("Starting to validate date")

	if data.EndDate != nil {
		if data.StartDate == nil {
			endDateValue := *data.EndDate
			if endDateValue.Before(sprint.StartDate) {
				logger.Error("End date must be after start date",
					"end_date", endDateValue,
					"start_date", sprint.StartDate)
			}
			return nil, structs.ErrEndDateBeforeStartDate
		}
	} else {
		if data.StartDate != nil {
			startDateValue := *data.StartDate
			if startDateValue.After(sprint.EndDate) {
				logger.Error("End date must be after start date",
					"start_date", startDateValue,
					"end_date", sprint.EndDate)
			}
			return nil, structs.ErrEndDateBeforeStartDate
		}
	}
	logger.Info("Date is valid")

	updateMap := make(map[string]any)
	if data.Name != nil {
		updateMap["name"] = *data.Name
	}
	if data.Goal != nil {
		updateMap["goal"] = *data.Goal
	}
	if data.StartDate != nil {
		updateMap["start_date"] = data.StartDate
	}
	if data.EndDate != nil {
		updateMap["end_date"] = data.EndDate
	}
	if len(updateMap) == 0 {
		logger.Info("No fields to update, returning current sprint")
		return sprint, nil
	}

	logger.Debug("Attempting project update operation", "input", updateMap)

	if err := s.sprintRepository.Update(ctx, sprintID, updateMap); err != nil {
		logger.Error("Failed to update sprint in repository", "error", err)
		return nil, structs.ErrDatabaseFail
	}

	logger.Info("Succesfully updated")

	updatedSprint, _ := s.sprintRepository.FindByID(ctx, sprintID)
	return updatedSprint, nil
}

func (s *sprintService) DeleteSprint(ctx context.Context, userID, sprintID int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "SprintService",
		"method", "DeleteSprint",
		"sprint_id", sprintID,
		"requestor_id", userID,
	)

	logger.Debug("Starting sprint deletion process")
	_, err := s.GetAndVerifyProjectManagerForSprint(ctx, logger, userID, sprintID)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotManageProject) {
			return fmt.Errorf("authorization failure for user id %d: %w", userID, err)
		} else {
			return fmt.Errorf("cannot fetch sprint: %w with sprint id: %d", err, sprintID)
		}
	}

	logger.Info("Authorization successful, attempting sprint deletion")

	if err := s.sprintRepository.Delete(ctx, sprintID); err != nil {
		logger.Error("Failed to delete sprint in repository", "error", err)
		return fmt.Errorf("repository delete failed for sprint %d: %w", sprintID, structs.ErrDatabaseFail)
	}

	logger.Info("Successfully deleted sprint")
	return nil
}
