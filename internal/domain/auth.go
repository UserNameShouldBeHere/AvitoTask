package domain

import (
	"fmt"

	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

type UserCredantials struct {
	UserName string
	Password string
}

func (userCredantialsLog *UserCredantials) Validate() error {
	if len(userCredantialsLog.UserName) < 3 ||
		len(userCredantialsLog.UserName) >= 150 {
		return fmt.Errorf("%w (Validate): incorrect name length", customErrors.ErrDataNotValid)
	}

	if len(userCredantialsLog.Password) < 6 {
		return fmt.Errorf("%w (Validate): incorrect password length", customErrors.ErrDataNotValid)
	}

	return nil
}
