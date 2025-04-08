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

type GenericRepository[T models.Identifiable[K], K comparable] struct {
	db          *gorm.DB
	modelName   string
	notFoundErr error
}

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

func (r *GenericRepository[T, K]) Create(ctx context.Context, model T) (T, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "GenericRepository",
		"method", "Create",
		"model", r.modelName,
	)
	logger.Debug("Starting generic create process")

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		logger.Error("Generic create failed", "error", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model, structs.ErrDataViolateConstraint
		}
		return model, fmt.Errorf("failed to create %s: %w", r.modelName, err)
	}

	logger.Info("Successfully created "+r.modelName, "id", model.GetID())
	return model, nil
}

func (r *GenericRepository[T, K]) FindByID(ctx context.Context, id K) (T, error) {
	var model T
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "GenericRepository",
		"method", "FindByID",
		"model", r.modelName,
		"id", id,
	)
	logger.Debug("Starting generic find by ID process")

	pkColumn := model.GetPKColumnName()
	if pkColumn == "" {
		err := errors.New("primary key column name cannot be empty")
		logger.Error("Configuration error", "error", err)
		return model, err
	}

	err := r.db.WithContext(ctx).First(&model,id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn(r.modelName + " not found")
			return model, r.notFoundErr
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
		return fmt.Errorf("failed to update %s %v: %w", r.modelName, id, result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("Generic update executed but no "+r.modelName+" found with the given ID or data was unchanged", "id", id)
		return r.notFoundErr
	}

	logger.Info("Successfully updated "+r.modelName, "id", id, "rows_affected", result.RowsAffected)
	return nil
}

func (r *GenericRepository[T, K]) Delete(ctx context.Context, id K) error {
	var model T
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

	result := r.db.WithContext(ctx).Delete(&model,id)

	if result.Error != nil {
		logger.Error("Generic delete failed", "error", result.Error)
		return structs.ErrDatabaseFail
	}

	if result.RowsAffected == 0 {
		logger.Warn("Generic delete executed but no "+r.modelName+" found with the given ID", "id", id)
		return r.notFoundErr
	}

	logger.Info("Successfully deleted "+r.modelName, "id", id, "rows_affected", result.RowsAffected)
	return nil
}