package repository

import (
	"context"
	"errors"
	"fmt"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/query"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByIDs(ctx context.Context, userIDs []int) ([]*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, id int, updateMap map[string]any) error
	Delete(ctx context.Context, id int) error
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
			return nil, structs.ErrEmailNotExist
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

func (r *userRepository) FindByIDs(ctx context.Context, userIDs []int) ([]*models.User, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserRepository",
		"method", "FindUsersByIDs",
		"userIDs", userIDs,
	)
	logger.Debug("Fetching users by IDs")

	if len(userIDs) == 0 {
		logger.Debug("No user IDs provided, returning empty list")
		return []*models.User{}, nil
	}

	u := r.q.User

	users, err := u.WithContext(ctx).Where(u.ID.In(userIDs...)).Find()
	if err != nil {
		logger.Error("Database query failed while fetching users by IDs", "error", err)
		return nil, err
	}

	logger.Debug("User query successful", "found_users_count", len(users))
	return users, nil
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
