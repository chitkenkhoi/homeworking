package handler

import (
    "lqkhoi-go-http-api/internal/dto"
)

func createErrorResponse(msg string, details interface{}) dto.ErrorResponse {
    return dto.ErrorResponse{
        Message: msg,
        Details: details,
    }
}

func createSuccessResponse[T any](msg string, data T) dto.SuccessResponse[T] {
    return dto.SuccessResponse[T]{
        Message: msg,
        Data:    data,
    }
}

func createSliceSuccessResponseGeneric[T any](msg string, data []T) dto.SliceSuccessResponse[T] {
    return dto.SliceSuccessResponse[T]{
        Message: msg,
        Data:    data,
        Count:   len(data),
    }
}