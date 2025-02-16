package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap/zaptest"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	"github.com/UserNameShouldBeHere/AvitoTask/internal/infrastructure/postgres"
	"github.com/UserNameShouldBeHere/AvitoTask/internal/services"
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

func TestAuthPostgres(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		"5432",
		"postgres",
		"root1234",
		"shop",
	))
	if err != nil {
		log.Fatalf("error in postgres initialization: %v\n", err)
	}

	logger := zaptest.NewLogger(t).Sugar()

	sessionExpiration := 60

	authStorage, err := postgres.NewAuthStorage(pool)
	if err != nil {
		log.Fatalf("error in auth storage initialization: %v\n", err)
	}

	authService, err := services.NewAuthService(authStorage, logger, 10, sessionExpiration)
	if err != nil {
		log.Fatalf("error in auth service initialization: %v\n", err)
	}

	authHandler, err := NewAuthHandler(authService, logger, sessionExpiration)
	if err != nil {
		log.Fatalf("error in auth handler initialization: %v\n", err)
	}

	authData := domain.UserCredantials{
		UserName: "test_user",
		Password: "test_password",
	}

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
