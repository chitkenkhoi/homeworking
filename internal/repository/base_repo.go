package repository

import (
	"context"
	"errors"
	"fmt"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"gorm.io/gorm"
)

// GenericRepository provides basic CRUD operations using standard GORM.
// T is the model type (e.g., *models.User), K is the primary key type (e.g., int).
// T must implement models.Identifiable[K].
type GenericRepository[T models.Identifiable[K], K comparable] struct {
	db          *gorm.DB
	modelName   string // For logging purposes
	notFoundErr error  // Specific "not found" error for this type
}

// NewGenericRepository creates a new generic repository instance.
func NewGenericRepository[T models.Identifiable[K], K comparable](
	db *gorm.DB,
	modelName string,
	notFoundErr error,
) *GenericRepository[T, K] {
	if notFoundErr == nil {
		notFoundErr = errors.New(modelName + " not found")
	}
	return &GenericRepository[T, K]{
		db:          db,
		modelName:   modelName,
		notFoundErr: notFoundErr,
	}
}

// Create uses standard GORM Create.
// NOTE: Loses gen's type safety for column names. Relies on GORM struct tags.
// NOTE: Basic constraint error handling. May need refinement.
func (r *GenericRepository[T, K]) Create(ctx context.Context, model T) (T, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "GenericRepository",
		"method", "Create",
		"model", r.modelName,
	)
	logger.Debug("Starting generic create process")

	// Use standard GORM Create
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		logger.Error("Generic create failed", "error", err)
		// Attempt basic constraint violation check (may vary by DB driver)
		// You might need a more robust error checking utility
		if errors.Is(err, gorm.ErrDuplicatedKey) { // Check for specific GORM v2 errors if applicable
			return model, structs.ErrDataViolateConstraint // Use your common constraint error
		}
		// Consider checking for other constraint types if needed
		return model, fmt.Errorf("failed to create %s: %w", r.modelName, err)
	}

	logger.Info("Successfully created "+r.modelName, "id", model.GetID())
	return model, nil
}

// FindByID uses standard GORM Where("pk = ?", id).First().
// NOTE: Loses gen's type safety for the Where clause. Relies on GetPKColumnName().
func (r *GenericRepository[T, K]) FindByID(ctx context.Context, id K) (T, error) {
	var model T // Must be the zero value of the pointer type if T is a pointer, or zero value of struct type
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "GenericRepository",
		"method", "FindByID",
		"model", r.modelName,
		"id", id,
	)
	logger.Debug("Starting generic find by ID process")

	// Get PK column name from the model instance (even zero value works)
	pkColumn := model.GetPKColumnName()
	if pkColumn == "" {
		err := errors.New("primary key column name cannot be empty")
		logger.Error("Configuration error", "error", err)
		return model, err // Return zero value of T and error
	}

	// Use standard GORM Find
	err := r.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", pkColumn), id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn(r.modelName + " not found")
			return model, r.notFoundErr // Return the specific not found error configured
		}
		logger.Error("Generic find by ID failed", "error", err)
		return model, fmt.Errorf("database error finding %s %v: %w", r.modelName, id, err)
	}

	logger.Info("Successfully found "+r.modelName+" by ID")
	return model, nil
}

func (r *GenericRepository[T, K]) Update(ctx context.Context, id K, updateMap map[string]any) error {
	var model T
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "GenericRepository",
		"method", "Update",
		"model", r.modelName,
		"id", id,
	)
	logger.Debug("Starting generic update process", "update_data_keys", utils.MapKeys(updateMap))

	if len(updateMap) == 0 {
		logger.Info("Update map is empty, skipping database call.")
		return nil
	}

	pkColumn := model.GetPKColumnName()
	if pkColumn == "" {
		err := errors.New("primary key column name cannot be empty")
		logger.Error("Configuration error", "error", err)
		return err
	}

	result := r.db.WithContext(ctx).Model(&model).Where(fmt.Sprintf("%s = ?", pkColumn), id).Updates(updateMap)

	if result.Error != nil {
		logger.Error("Generic update failed", "error", result.Error)
		// Otherwise, return a wrapped error
		return fmt.Errorf("failed to update %s %v: %w", r.modelName, id, result.Error)
	}

	// If RowsAffected is 0, GORM might mean "record not found" OR "data was identical, no change needed".
	// Following the pattern of the original specific repository, we treat 0 rows affected as if the record wasn't found.
	if result.RowsAffected == 0 {
		logger.Warn("Generic update executed but no "+r.modelName+" found with the given ID or data was unchanged", "id", id)
		return r.notFoundErr // Return the specific not found error
	}

	logger.Info("Successfully updated "+r.modelName, "id", id, "rows_affected", result.RowsAffected)
	return nil
}
// Delete uses standard GORM Where("pk = ?", id).Delete().
// NOTE: Loses gen's type safety for the Where clause. Relies on GetPKColumnName().
func (r *GenericRepository[T, K]) Delete(ctx context.Context, id K) error {
	var model T // Needed for GORM Delete signature and getting PK column
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "GenericRepository",
		"method", "Delete",
		"model", r.modelName,
		"id", id,
	)
	logger.Debug("Starting generic delete process")

	pkColumn := model.GetPKColumnName()
	if pkColumn == "" {
		err := errors.New("primary key column name cannot be empty")
		logger.Error("Configuration error", "error", err)
		return err
	}

	// Use standard GORM Delete
	// Pass address of the zero value model type
	result := r.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", pkColumn), id).Delete(&model)

	if result.Error != nil {
		logger.Error("Generic delete failed", "error", result.Error)
		// Return a generic database failure error or wrap
		return structs.ErrDatabaseFail
	}

	if result.RowsAffected == 0 {
		logger.Warn("Generic delete executed but no "+r.modelName+" found with the given ID", "id", id)
		return r.notFoundErr // Return the specific not found error
	}

	logger.Info("Successfully deleted "+r.modelName, "id", id, "rows_affected", result.RowsAffected)
	return nil
}