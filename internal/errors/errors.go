package errors

import (
	"errors"
	"net/http"
)

var (
	ErrInternal                 = errors.New("internal service error")
	ErrDataNotValid             = errors.New("invalid data")
	ErrIncorrectEmailOrPassword = errors.New("incorrect email or password")
)

func ConvertToHttpErr(err error) int {
	switch {
	case errors.Is(err, ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, ErrIncorrectEmailOrPassword),
		errors.Is(err, ErrDataNotValid),
		errors.Is(err, ErrDoesNotExist):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
