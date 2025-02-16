package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	serviceMocks "github.com/UserNameShouldBeHere/AvitoTask/internal/services/mocks"
)

func TestAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := serviceMocks.NewMockAuthService(ctrl)

	logger := zaptest.NewLogger(t).Sugar()

	sessionExpiration := 60

	authHandler, err := NewAuthHandler(authService, logger, sessionExpiration)
	if err != nil {
		log.Fatalf("error in shop handler initialization: %v\n", err)
	}

	authData := domain.UserCredantials{
		UserName: "test_user",
		Password: "test_password",
	}

	ctx := context.WithValue(context.Background(), CtxSessionName, authData.UserName)
	authService.EXPECT().LoginOrCreateUser(ctx, authData).Return("token", nil)

	jsonData, err := json.Marshal(authData)
	if err != nil {
		t.Error(err)
	}

	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(jsonData))

	authHandler.Auth(wr, req)
	if wr.Code != http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}

	authData = domain.UserCredantials{
		UserName: "",
		Password: "test_password",
	}

	jsonData, err = json.Marshal(authData)
	if err != nil {
		t.Error(err)
	}

	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(jsonData))

	authHandler.Auth(wr, req)
	if wr.Code != http.StatusBadRequest {
		t.Errorf("got HTTP status code %d, expected 400", wr.Code)
	}

	authData = domain.UserCredantials{
		UserName: "test_usesr",
		Password: "",
	}

	jsonData, err = json.Marshal(authData)
	if err != nil {
		t.Error(err)
	}

	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(jsonData))

	authHandler.Auth(wr, req)
	if wr.Code != http.StatusBadRequest {
		t.Errorf("got HTTP status code %d, expected 400", wr.Code)
	}
}
