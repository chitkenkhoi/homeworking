package custom_structs
import (
	"errors"
)

var (
	ErrInternalServer = errors.New("internal server error")
	ErrDatabaseFail      = errors.New("internal database fail")
	ErrPasswordIncorrect = errors.New("password is incorrect")
	ErrUsernameOrEmailNotExist = errors.New("username or email does not exist")
	ErrTokenCanNotBeSigned = errors.New("can not signed the token")	
)
