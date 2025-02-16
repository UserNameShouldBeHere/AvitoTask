package services

import (
	"context"
	"fmt"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	"go.uber.org/zap"
)

type ShopStorage interface {
	GetInfo(ctx context.Context, username string) (domain.InventoryInfo, error)
	SendCoin(ctx context.Context, transaction domain.Transaction) error
	BuyItem(ctx context.Context, username string, itemName string) error
}

type ShopService struct {
	shopStorage ShopStorage
	logger      *zap.SugaredLogger
}

func NewShopService(shopStorage ShopStorage, logger *zap.SugaredLogger) (*ShopService, error) {
	return &ShopService{
		shopStorage: shopStorage,
		logger:      logger,
	}, nil
}

func (shopService *ShopService) GetInfo(ctx context.Context, username string) (domain.InventoryInfo, error) {
	info, err := shopService.shopStorage.GetInfo(ctx, username)
	if err != nil {
		shopService.logger.Errorf("failed to get user info (service.GetInfo): %w", err)
		return domain.InventoryInfo{}, fmt.Errorf("(service.GetInfo): %w", err)
	}

	return info, nil
}

func (shopService *ShopService) SendCoin(ctx context.Context, transaction domain.Transaction) error {
	err := shopService.shopStorage.SendCoin(ctx, transaction)
	if err != nil {
		shopService.logger.Errorf("failed to send coins (service.SendCoin): %w", err)
		return fmt.Errorf("(service.SendCoin): %w", err)
	}

	return nil
}

func (shopService *ShopService) BuyItem(ctx context.Context, username string, itemName string) error {
	err := shopService.shopStorage.BuyItem(ctx, username, itemName)
	if err != nil {
		shopService.logger.Errorf("failed to buy item (service.BuyItem): %w", err)
		return fmt.Errorf("(service.BuyItem): %w", err)
	}

	return nil
}
