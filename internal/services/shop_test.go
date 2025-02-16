package services

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
	storageMocks "github.com/UserNameShouldBeHere/AvitoTask/internal/infrastructure/mocks"
)

func TestGetInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shopStorage := storageMocks.NewMockShopStorage(ctrl)

	logger := zaptest.NewLogger(t).Sugar()

	shopService, err := NewShopService(shopStorage, logger)
	if err != nil {
		log.Fatalf("error in shop handler initialization: %v\n", err)
	}

	info := domain.InventoryInfo{}

	testData := []struct {
		TestName string
		UserName string
		Error    error
	}{
		{
			"correct data",
			"test_user",
			nil,
		},
		{
			"incorrect user name",
			"unknown_user",
			customErrors.ErrDoesNotExist,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.TestName, func(t *testing.T) {
			shopStorage.EXPECT().GetInfo(context.Background(), testCase.UserName).Return(info, testCase.Error)

			_, err = shopService.GetInfo(context.Background(), testCase.UserName)
			if !errors.Is(err, testCase.Error) {
				t.Error(err)
			}
		})
	}
}

func TestSendCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shopStorage := storageMocks.NewMockShopStorage(ctrl)

	logger := zaptest.NewLogger(t).Sugar()

	shopService, err := NewShopService(shopStorage, logger)
	if err != nil {
		log.Fatalf("error in shop handler initialization: %v\n", err)
	}

	testData := []struct {
		TestName    string
		Transaction domain.Transaction
		Error       error
	}{
		{
			"correct data",
			domain.Transaction{
				From:   "test_user_1",
				To:     "test_user_2",
				Amount: 100,
			},
			nil,
		},
		{
			"unkown user",
			domain.Transaction{
				From:   "unknown_user",
				To:     "test_user_2",
				Amount: 100,
			},
			customErrors.ErrDoesNotExist,
		},
		{
			"incorrect amount",
			domain.Transaction{
				From:   "test_user_1",
				To:     "test_user_2",
				Amount: -1,
			},
			customErrors.ErrDataNotValid,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.TestName, func(t *testing.T) {
			shopStorage.EXPECT().SendCoin(context.Background(), testCase.Transaction).Return(testCase.Error)

			err = shopService.SendCoin(context.Background(), testCase.Transaction)
			if !errors.Is(err, testCase.Error) {
				t.Error(err)
			}
		})
	}
}

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shopStorage := storageMocks.NewMockShopStorage(ctrl)

	logger := zaptest.NewLogger(t).Sugar()

	shopService, err := NewShopService(shopStorage, logger)
	if err != nil {
		log.Fatalf("error in shop handler initialization: %v\n", err)
	}

	testData := []struct {
		TestName string
		UserName string
		ItemName string
		Error    error
	}{
		{
			"correct data",
			"test_user",
			"t-shirt",
			nil,
		},
		{
			"incorrect item",
			"test_user",
			"blade",
			customErrors.ErrDoesNotExist,
		},
		{
			"user has less money than item price",
			"test_user",
			"t-shirt",
			customErrors.ErrDataNotValid,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.TestName, func(t *testing.T) {
			shopStorage.EXPECT().BuyItem(context.Background(), testCase.UserName, testCase.ItemName).Return(testCase.Error)

			err = shopService.BuyItem(context.Background(), testCase.UserName, testCase.ItemName)
			if !errors.Is(err, testCase.Error) {
				t.Error(err)
			}
		})
	}
}
