package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type AuthService interface {
	LoginOrCreateUser(ctx context.Context, userCreds domain.UserCredantials) (string, error)
	GetNameAndCheck(ctx context.Context, token string) (string, bool)
}

type AuthHandler struct {
	authService       AuthService
	logger            *zap.SugaredLogger
	sessionExpiration int
}

func NewAuthHandler(authService AuthService, logger *zap.SugaredLogger, sessionExpiration int) (*AuthHandler, error) {
	return &AuthHandler{
		authService:       authService,
		logger:            logger,
		sessionExpiration: sessionExpiration,
	}, nil
}

func (h *AuthHandler) Auth(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var userCreds domain.UserCredantials
	err = json.Unmarshal(body, &userCreds)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	if err = userCreds.Validate(); err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to write http response: %v", err)
		}
		return
	}

	ctx := context.WithValue(req.Context(), CtxSessionName, userCreds.UserName)

	token, err := h.authService.LoginOrCreateUser(ctx, userCreds)
	if err != nil {
		err = WriteResponse(w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to write http response: %v", err)
		}
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/api",
		MaxAge:   h.sessionExpiration,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	err = WriteResponse(
		w,
		h.logger,
		ResponseData{
			Session: token,
			Url:     req.Pattern,
			Status:  http.StatusOK,
			Data:    TokenResponse{Token: token},
		})
	if err != nil {
		h.logger.Errorf("unable to write http response: %v", err)
	}
}
