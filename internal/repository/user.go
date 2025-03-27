package repository

import (
	"errors"
	"log/slog"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	FindByID(id int) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	List() ([]models.User, error)
	Update(user *models.User) error
	Delete(id int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) (*models.User, error) {
	if err := r.db.Create(user).Error; err != nil {
		slog.Error("Can not create user", "error", err)
		return nil, structs.ErrDataViolateConstraint
	}
	return user, nil
}

func (r *userRepository) FindByID(id int) (*models.User, error) {
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

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List() ([]models.User, error) {
	var users []models.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id int) error {
	tx := r.db.Delete(&models.User{}, id)
	err := tx.Error
	if err != nil {
		slog.Error("Internal database fail", "error", err)
		return err
	}

	if tx.RowsAffected == 0 {
		return structs.ErrUserNotExist
	}
	return nil
}
