package utils

import (
	"log"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

func ValidateStruct(payload any) []*ErrorResponse {
	var errs []*ErrorResponse
	validate := validator.New()
	err := validate.Struct(payload)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			log.Printf("Unexpected validation error type: %T\n", err)
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
