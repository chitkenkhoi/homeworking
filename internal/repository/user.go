package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/query"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, id int, updateMap map[string]any) error
	Delete(ctx context.Context, id int) error
	FindValidTeamMembersForAssignment(ctx context.Context, userIDs []int) ([]int, error)
	AssignUsersToProject(ctx context.Context, projectID int, userIDs []int) error
}

type userRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,            // Store the original db
		q:  query.Use(db), // <--- Initialize the query object here
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "Create",
	)
	logger.Debug("Starting create user process", "email", user.Email, "role", user.Role)

	err := r.q.User.WithContext(ctx).Create(user)
	if err != nil {
		logger.Error("Failed to create user", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}

	logger.Info("Successfully created user", "user_id", user.ID)
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "FindByID",
		"user_id", id,
	)
	logger.Debug("Starting find user by ID process")

	u := r.q.User
	// Use generated Where and First
	user, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("User not found")
			return nil, structs.ErrUserNotExist
		}
		logger.Error("Failed to find user by ID due to database error", "error", err)
		return nil, fmt.Errorf("database error finding user %d: %w", id, err)
	}

	logger.Info("Successfully found user by ID")
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "FindByEmail",
		"email", email,
	)
	logger.Debug("Starting find user by email process")

	u := r.q.User
	user, err := u.WithContext(ctx).Where(u.Email.Eq(email)).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("User with email not found")
			return nil, structs.ErrUserNotExist
		}
		logger.Error("Failed to find user by email due to database error", "error", err)
		return nil, fmt.Errorf("database error finding user by email %s: %w", email, err)
	}

	logger.Info("Successfully found user by email", "user_id", user.ID)
	return user, nil
}

func (r *userRepository) List(ctx context.Context) ([]*models.User, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "List",
	)
	logger.Debug("Starting list users process")

	users, err := r.q.User.WithContext(ctx).Find()
	if err != nil {
		logger.Error("Failed to list users due to database error", "error", err)
		return nil, fmt.Errorf("database error listing users: %w", err)
	}

	logger.Info("Successfully listed users", "count", len(users))
	logger.Debug("Listed users details", "users", users)
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id int, updateMap map[string]any) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "Update",
		"user_id", id,
	)
	logger.Debug("Starting update user process", "update_data_keys", utils.MapKeys(updateMap)) // Log only keys

	if len(updateMap) == 0 {
		logger.Info("Update map is empty after validation, skipping database call.")
		return nil
	}

	u := r.q.User
	resultInfo, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Updates(updateMap)

	if err != nil {
		logger.Error("Failed to update user", "error", err)
		return fmt.Errorf("failed to update user %d: %w", id, err)
	}

	if resultInfo.RowsAffected == 0 {
		logger.Warn("Update executed but no user found with the given ID or data was the same")
		return structs.ErrUserNotExist
	}

	logger.Info("Successfully updated user", "rows_affected", resultInfo.RowsAffected)
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "Delete",
		"user_id", id,
	)
	logger.Debug("Starting delete user process")

	u := r.q.User
	resultInfo, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()

	if err != nil {
		logger.Error("Failed to delete user due to database error", "error", err)
		return structs.ErrDatabaseFail
	}

	if resultInfo.RowsAffected == 0 {
		logger.Warn("Delete executed but no user found with the given ID")
		return structs.ErrUserNotExist
	}

	logger.Info("Successfully deleted user", "rows_affected", resultInfo.RowsAffected)
	return nil
}

func (r *userRepository) FindValidTeamMembersForAssignment(ctx context.Context, userIDs []int) ([]int, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "FindValidTeamMembersForAssignment",
		"userIDs", userIDs,
	)
	logger.Debug("Finding valid team members for assignment")

	u := r.q.User

	users, err := u.WithContext(ctx).Where(u.ID.In(userIDs...)).Find()
	if err != nil {
		logger.Error("Database query failed while fetching users", "error", err)
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	logger.Debug("Query successful", "found_users_count", len(users))

	invalidUserMessages := make([]string, 0, len(userIDs))

	if len(users) != len(userIDs) {
		foundIDs := make(map[int]struct{}, len(users))
		for _, u := range users {
			foundIDs[u.ID] = struct{}{}
		}
		var missing []int
		for _, reqID := range userIDs {
			if _, ok := foundIDs[reqID]; !ok {
				missing = append(missing, reqID)
				invalidUserMessages = append(invalidUserMessages, fmt.Sprintf("user %d not found", reqID))
			}
		}
		logger.Error("Some requested users not found", "missing_ids", missing)
	}

	var validUserIDs []int
	for _, user := range users {
		userLogger := logger.With("user_id", user.ID)
		userLogger.Debug("Validating user eligibility")
		if user.Role != models.TeamMember {
			msg := fmt.Sprintf("user %d has incorrect role '%s' (required: '%s')", user.ID, user.Role, models.TeamMember)
			userLogger.Warn("Invalid role for assignment", "current_role", user.Role)
			invalidUserMessages = append(invalidUserMessages, msg)
			continue
		}
		if user.CurrentProjectID != nil {
			msg := fmt.Sprintf("user %d is already assigned to project %d", user.ID, *user.CurrentProjectID)
			userLogger.Warn("User already assigned to a project", "project_id", *user.CurrentProjectID)
			invalidUserMessages = append(invalidUserMessages, msg)
			continue
		}
		userLogger.Debug("User is eligible for assignment")
		validUserIDs = append(validUserIDs, user.ID)
	}

	if len(invalidUserMessages) > 0 {
		joinedMessages := strings.Join(invalidUserMessages, "; ")
		logger.Warn("Some users failed validation for assignment", "fail_count", len(invalidUserMessages), "errors", joinedMessages)
		return validUserIDs, fmt.Errorf("validation failed for some users: %s", joinedMessages)
	}

	logger.Info("All requested users validated successfully for assignment", "valid_count", len(validUserIDs))
	return validUserIDs, nil
}

func (r *userRepository) AssignUsersToProject(ctx context.Context, projectID int, userIDs []int) (err error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "AssignUsersToProject",
		"project_id", projectID,
		"user_ids", userIDs,
	)
	logger.Debug("Starting user assignment to project in transaction", "user_count", len(userIDs))

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		logger.Error("Failed to begin database transaction", "error", tx.Error)
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic recovered during transaction, rolling back", "panic_value", r)
			tx.Rollback()
			panic(r)
		} else if err != nil {
			logger.Warn("Rolling back transaction due to error", "error", err)
			if rbErr := tx.Rollback().Error; rbErr != nil {
				logger.Error("Failed to rollback transaction after error", "rollback_error", rbErr, "original_error", err)
			}
		} else {
			logger.Debug("Transaction committed successfully or skipped (no users).")
		}
	}()

	qTx := query.Use(tx)
	u := qTx.User

	updateData := map[string]interface{}{
		u.CurrentProjectID.ColumnName().String(): projectID,
	}

	resultInfo, updateErr := u.WithContext(ctx).Where(u.ID.In(userIDs...)).Updates(updateData)

	if updateErr != nil {
		logger.Error("Failed to update users' project assignment within transaction", "error", updateErr)
		err = fmt.Errorf("failed to update users' project assignment: %w", updateErr)
		return err
	}

	if resultInfo.RowsAffected != int64(len(userIDs)) {
		errMsg := fmt.Sprintf("unexpected number of users updated: expected %d, got %d", len(userIDs), resultInfo.RowsAffected)
		logger.Error("User assignment row count mismatch", "expected", len(userIDs), "actual", resultInfo.RowsAffected)
		err = errors.New(errMsg)
		return err
	}

	logger.Debug("Users updated successfully within transaction, attempting commit", "rows_affected", resultInfo.RowsAffected)

	if commitErr := tx.Commit().Error; commitErr != nil {
		logger.Error("Failed to commit transaction", "error", commitErr)
		err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		return err
	}

	logger.Info("Successfully assigned users to project and committed transaction")
	return nil
}
