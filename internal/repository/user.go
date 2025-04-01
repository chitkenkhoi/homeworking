package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
	FindValidTeamMembersForAssignment(ctx context.Context, userIDs []int) ([]int, error)
	AssignUsersToProject(ctx context.Context, projectID int, userIDs []int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.db.Create(user).Error; err != nil {
		slog.Error("Can not create user", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, structs.ErrUserNotExist
		}
		slog.Error("Internal database failed", "err", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "Delete",
		"user_id", id,
	)

	logger.Debug("Starting delete user process")

	tx := r.db.Delete(&models.User{}, id)
	err := tx.Error
	if err != nil {
		logger.Error("Internal database fail", "error", err)
		return structs.ErrDatabaseFail
	}

	if tx.RowsAffected == 0 {
		logger.Error("User with this id does not exist")
		return structs.ErrUserNotExist
	}

	logger.Info("User is deleted")

	return nil
}

func (r *userRepository) FindValidTeamMembersForAssignment(ctx context.Context, userIDs []int) ([]int, error) {
	logger := slog.With("method", "FindValidTeamMembersForAssignment", "userIDs", userIDs)
	logger.Debug("Finding valid team members for assignment")

	var users []models.User

	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		logger.Error("Database query failed", "error", err)
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	logger.Debug("Query successful", "found_users_count", len(users))

	if len(users) != len(userIDs) {
		foundIDs := make(map[int]struct{}, len(users))
		for _, u := range users {
			foundIDs[u.ID] = struct{}{}
		}
		var missing []int
		for _, reqID := range userIDs {
			if _, ok := foundIDs[reqID]; !ok {
				missing = append(missing, reqID)
			}
		}

		logger.Error("Some requested users not found", "missing_ids", missing)
		return nil, fmt.Errorf("users not found: %v", missing)
	}

	var invalidUserMessages []string
	var validUserIDs []int
	for _, user := range users {
		userLogger := logger.With("user_id", user.ID)
		userLogger.Debug("Validating user eligibility")
		if user.Role != models.TeamMember {
			userLogger.Error("Invalid role",
				"current_role", user.Role,
				"required_role", models.TeamMember)
			invalidUserMessages = append(invalidUserMessages, fmt.Sprintf("user %d has incorrect role '%s'", user.ID, user.Role))
			continue
		}
		if user.CurrentProjectID != nil {
			userLogger.Error("User already assigned to project",
				"project_id", *user.CurrentProjectID)
			invalidUserMessages = append(invalidUserMessages, fmt.Sprintf("user %d is already assigned to project %d", user.ID, *user.CurrentProjectID))
			continue
		}
		userLogger.Debug("User is eligible for assignment")
		validUserIDs = append(validUserIDs, user.ID)
	}

	if len(invalidUserMessages) > 0 {
		logger.Error("Some users failed validation",
			"valid_count", len(validUserIDs),
			"invalid_count", len(invalidUserMessages))
		joinedMessages := strings.Join(invalidUserMessages, "; ")
		return validUserIDs, fmt.Errorf("validation failed for some users: %s", joinedMessages)
	}

	logger.Debug("All users validated successfully", "valid_count", len(validUserIDs))
	return validUserIDs, nil
}

func (r *userRepository) AssignUsersToProject(ctx context.Context, projectID int, userIDs []int) error {
	logger := slog.With("method", "AssignUsersToProject", "project_id", projectID, "user_ids", userIDs)

	logger.Debug("Starting user assignment", "user_count", len(userIDs))

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		logger.Error("Failed to begin database transaction", "error", tx.Error)
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	logger.Debug("Beginning database transaction")

	result := tx.Model(&models.User{}).Where("id IN ?", userIDs).Updates(map[string]any{
		"current_project_id": projectID,
	})

	if result.Error != nil {
		logger.Error("Failed to update users' project assignment", "error", result.Error)
		logger.Debug("Rolling back transaction due to update error")
		tx.Rollback()
		return fmt.Errorf("failed to update users' project assignment: %w", result.Error)
	}

	if result.RowsAffected != int64(len(userIDs)) {
		logger.Error("Unexpected number of users updated",
			"expected", len(userIDs),
			"actual", result.RowsAffected)
		logger.Debug("Rolling back transaction due to row count mismatch")
		tx.Rollback()
		return fmt.Errorf("unexpected number of users updated: expected %d, got %d", len(userIDs), result.RowsAffected)
	}

	logger.Debug("Attempting to commit transaction")
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		logger.Debug("Rolling back transaction due to commit error")
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
