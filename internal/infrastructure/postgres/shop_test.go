package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

func TestGetInfo(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage, err := NewShopStorage(mock)
	require.NoError(t, err)

	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})

	userName := "test_user"
	userId := 1
	userMoney := 1000

	mockRows := pgxmock.NewRows([]string{"id", "money"}).AddRow(userId, userMoney)

	mock.ExpectQuery("select").
		WithArgs(userName).
		WillReturnRows(mockRows)

	productName := "t-shirt"
	productCount := 1

	mockRows = pgxmock.NewRows([]string{"name", "count(*)"}).AddRow(productName, productCount)

	mock.ExpectQuery("select").
		WithArgs(userId).
		WillReturnRows(mockRows)

	anotherUser := "test_2_user"
	amount := 100

	mockRows = pgxmock.NewRows([]string{"name", "money"}).AddRow(anotherUser, amount)

	mock.ExpectQuery("select").
		WithArgs(userId).
		WillReturnRows(mockRows)

	mockRows = pgxmock.NewRows([]string{"name", "money"}).AddRow(anotherUser, amount)

	mock.ExpectQuery("select").
		WithArgs(userId).
		WillReturnRows(mockRows)

	mock.ExpectCommit()

	_, err = storage.GetInfo(context.Background(), userName)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestSendMoney(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage, err := NewShopStorage(mock)
	require.NoError(t, err)

	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})

	transaction := domain.Transaction{
		From:   "test_user",
		To:     "test_2_user",
		Amount: 100,
	}

	toUserId := 2

	mockRows := pgxmock.NewRows([]string{"id"}).AddRow(toUserId)

	mock.ExpectQuery("select").
		WithArgs(transaction.To).
		WillReturnRows(mockRows)

	fromUserId := 1
	userMoney := 1000

	mockRows = pgxmock.NewRows([]string{"id", "money"}).AddRow(fromUserId, userMoney)

	mock.ExpectQuery("select").
		WithArgs(transaction.From).
		WillReturnRows(mockRows)

	mock.ExpectExec("update").
		WithArgs(-transaction.Amount, fromUserId).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("update").
		WithArgs(transaction.Amount, toUserId).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("insert").
		WithArgs(fromUserId, toUserId, transaction.Amount).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	err = storage.SendCoin(context.Background(), transaction)
	require.NoError(t, err)

	transaction.Amount = -1
	err = storage.SendCoin(context.Background(), transaction)
	require.Error(t, err, customErrors.ErrDataNotValid)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestBuyItem(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage, err := NewShopStorage(mock)
	require.NoError(t, err)

	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})

	itemName := "t-shirt"
	itemId := 1
	itemPrice := 80

	mockRows := pgxmock.NewRows([]string{"id", "price"}).AddRow(itemId, itemPrice)

	mock.ExpectQuery("select").
		WithArgs(itemName).
		WillReturnRows(mockRows)

	userName := "test_user"
	userId := 1
	userMoney := 80

	mockRows = pgxmock.NewRows([]string{"id", "money"}).AddRow(userId, userMoney)

	mock.ExpectQuery("select").
		WithArgs(userName).
		WillReturnRows(mockRows)

	mock.ExpectExec("update").
		WithArgs(-itemPrice, userId).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("insert").
		WithArgs(userId, itemId).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	err = storage.BuyItem(context.Background(), userName, itemName)
	require.NoError(t, err)

	err = storage.BuyItem(context.Background(), userName, itemName)
	require.Error(t, err, customErrors.ErrDataNotValid)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
