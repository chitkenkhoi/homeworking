package dto

// SuccessResponse represents a generic success response
type SuccessResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Message string `json:"message" example:"An error occurred"`
	Details any    `json:"details,omitempty"`
}

// SliceSuccessResponse represents a success response with a slice of data
type SliceSuccessResponse[T any] struct {
	Message string `json:"message"`
	Data    []T    `json:"data"`
	Count   int    `json:"count" example:"3"`
}

// AddTeamMembersPartialSuccessResponse represents a partial success response for AddTeamMembers
type AddTeamMembersPartialSuccessResponse struct {
	Message      string `json:"message" example:"Some users could not be added to the project"`
	Details      string `json:"details" example:"User with ID 3 not found"`
	UpdatedCount int    `json:"updated_count" example:"2"`
}

// TokenResponse represents a response contain token for authentication
type TokenResponse struct {
	Token string `json:"token" example:"random-token"`
}

type ProjectSuccessResponse struct {
	Message string          `json:"message" example:"Operation successful"`
	Data    ProjectResponse `json:"data"`
}

type ProjectSliceSuccessResponse struct {
	Message string            `json:"message" example:"Items found successfully"`
	Data    []ProjectResponse `json:"data"`
	Count   int               `json:"count" example:"5"`
}

type GenericSuccessResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

type IntSuccessResponse struct {
	Message string `json:"message" example:"Operation successful"`
	Data    int    `json:"data" example:"3"`
}

type UserSuccessResponse struct {
	Message string       `json:"message" example:"Operation successful"`
	Data    UserResponse `json:"data"`
}

type UserSliceSuccessResponse struct {
	Message string         `json:"message" example:"Items found successfully"`
	Data    []UserResponse `json:"data"`
	Count   int            `json:"count" example:"5"`
}

type TaskSuccessResponse struct {
	Message string       `json:"message" example:"Operation successful"`
	Data    TaskResponse `json:"data"`
}

type TaskSliceSuccessResponse struct {
	Message string         `json:"message" example:"Items found successfully"`
	Data    []TaskResponse `json:"data"`
	Count   int            `json:"count" example:"5"`
}

type SprintSuccessResponse struct {
	Message string         `json:"message" example:"Operation successful"`
	Data    SprintResponse `json:"data"`
}

type SprintSliceSuccessResponse struct {
	Message string           `json:"message" example:"Items found successfully"`
	Data    []SprintResponse `json:"data"`
	Count   int              `json:"count" example:"5"`
}
