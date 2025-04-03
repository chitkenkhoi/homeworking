package utils

import (
	"log/slog"
	"sync"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

var (
	validate *validator.Validate
	once     sync.Once
)

func initValidator() {
	validate = validator.New()
}

func ValidateStruct(payload any) []*ErrorResponse {
	once.Do(initValidator)
	var errs []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			slog.Error("Unexpected validation error type", "error", err)
			return errs
		}

		for _, err := range validationErrors {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errs = append(errs, &element)
		}
	}
	return errs
}

func ValidateStructForConfig(payload any) error {
	once.Do(initValidator)
	err := validate.Struct(payload)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			slog.Error("Unexpected validation error type", "error", err)
			return err
		}

		for _, err := range validationErrors {
			slog.Error("Config invalid", "error", err)
		}
		return err
	}
	slog.Info("Config are valid")
	return nil
}
