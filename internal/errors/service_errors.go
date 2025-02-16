package errors

import "errors"

var (
	ErrFailedToGenJWTKey     = errors.New("failed to generate key")
	ErrFailedToCreateToken   = errors.New("failed to create token")
	ErrFailedToExecuteMethod = errors.New("failed to execute method")
	ErrFailedToSignToken     = errors.New("failed to sign token")
	ErrUnauthenticated       = errors.New("unauthenticated")
)
