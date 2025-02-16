package postgres

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

func TestCreateUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage, err := NewAuthStorage(mock)
	require.NoError(t, err)

	userCreds := domain.UserCredantials{
		UserName: "test_user",
		Password: "test_password",
	}

	mock.ExpectQuery("select").
		WithArgs(userCreds.UserName).
		WillReturnRows(pgxmock.NewRows([]string{}))

	mock.ExpectExec("insert").
		WithArgs(userCreds.UserName, userCreds.Password).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = storage.CreateUser(context.Background(), userCreds)
	require.NoError(t, err)

	err = storage.CreateUser(context.Background(), userCreds)
	require.Error(t, err, customErrors.ErrAlreadyExists)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestGetPassword(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage, err := NewAuthStorage(mock)
	require.NoError(t, err)

	userName := "test_user"
	userPassword := "test_password"

	mockRows := pgxmock.NewRows([]string{"password"}).AddRow(userPassword)

	mock.ExpectQuery("select").
		WithArgs(userName).
		WillReturnRows(mockRows)

	_, err = storage.GetPassword(context.Background(), userName)
	require.NoError(t, err)

	userName = "unknown_user"

	_, err = storage.GetPassword(context.Background(), userName)
	require.Error(t, err, customErrors.ErrDoesNotExist)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestHasUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage, err := NewAuthStorage(mock)
	require.NoError(t, err)

	userName := "test_user"

	mock.ExpectQuery("select").
		WithArgs(userName).
		WillReturnRows(pgxmock.NewRows([]string{}))

	_, err = storage.HasUser(context.Background(), userName)
	require.NoError(t, err)

	userName = "unknown_user"

	mock.ExpectQuery("select").
		WithArgs(userName).
		WillReturnRows(pgxmock.NewRows([]string{}))

	ok, err := storage.HasUser(context.Background(), userName)
	require.NoError(t, err)
	require.Equal(t, false, ok)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
