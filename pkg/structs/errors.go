package structs

import (
	"errors"
)

var (
	ErrInternalServer         = errors.New("internal server error")
	ErrDatabaseFail           = errors.New("internal database fail")
	ErrPasswordIncorrect      = errors.New("password is incorrect")
	ErrEmailNotExist          = errors.New("email does not exist")
	ErrTokenCanNotBeSigned    = errors.New("can not signed the token")
	ErrPasswordTooLong        = errors.New("password is too long")
	ErrRedisKeyNotExist       = errors.New("redis key does not exist")
	ErrRedisConnection        = errors.New("redis query failed")
	ErrUserNotExist           = errors.New("user does not exist")
	ErrDataViolateConstraint  = errors.New("data violate database constraints")
	ErrProjectNotExist        = errors.New("project does not exist")
	ErrUserNotManageProject   = errors.New("user does not manage this project")
	ErrNoValidUserStatus      = errors.New("no user has valid status")
	ErrEndDateBeforeStartDate = errors.New("end date must be after start date")
	ErrSprintDateInvalid      = errors.New("sprint date is invalid")
	ErrSprintNotExist         = errors.New("sprint does not exist")
	ErrTaskNotExist           = errors.New("task does not exist")
)
