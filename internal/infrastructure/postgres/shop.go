package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

type ShopStorage struct {
	pool PgxPool
}

func NewShopStorage(pool PgxPool) (*ShopStorage, error) {
	return &ShopStorage{
		pool: pool,
	}, nil
}

func (shopStorage *ShopStorage) GetInfo(ctx context.Context, username string) (domain.InventoryInfo, error) {
	tx, err := shopStorage.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrFailedToBeginTx, err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			fmt.Printf("%v (postgres.GetInfo): %v", customErrors.ErrFailedToRollbackTx, err)
		}
	}()

	var (
		userId    int
		userMoney int
	)
	err = tx.QueryRow(ctx, `
		select id, money
		from users
		where name = $1;
	`, username).Scan(&userId, &userMoney)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrDoesNotExist, err)
		}

		return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	inventory, err := shopStorage.getInventory(ctx, tx, userId)
	if err != nil {
		return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	recievedCoins, err := shopStorage.getRecievedCoins(ctx, tx, userId)
	if err != nil {
		return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	sentCoins, err := shopStorage.getSentCoins(ctx, tx, userId)
	if err != nil {
		return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	inventoryInfo := domain.InventoryInfo{
		Coins:     userMoney,
		Inventory: inventory,
		CoinHistory: domain.SentRecievedHistory{
			Recieved: recievedCoins,
			Sent:     sentCoins,
		},
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.InventoryInfo{}, fmt.Errorf("%w (postgres.GetInfo): %w", customErrors.ErrFailedToCommitTx, err)
	}

	return inventoryInfo, nil
}

func (shopStorage *ShopStorage) SendCoin(ctx context.Context, transaction domain.Transaction) error {
	tx, err := shopStorage.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToBeginTx, err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			fmt.Printf("%v (postgres.SendCoin): %v", customErrors.ErrFailedToRollbackTx, err)
		}
	}()

	var toUserId int
	err = tx.QueryRow(ctx, `
		select id
		from users
		where name = $1;
	`, transaction.To).Scan(&toUserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrDoesNotExist, err)
		}

		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	var (
		fromUserId int
		userMoney  int
	)
	err = tx.QueryRow(ctx, `
		select id, money
		from users
		where name = $1;
	`, transaction.From).Scan(&fromUserId, &userMoney)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrDoesNotExist, err)
		}

		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	if userMoney-transaction.Amount < 0 {
		return fmt.Errorf("%w (postgres.SendCoin): incorrect amount of coins", customErrors.ErrDataNotValid)
	}

	err = shopStorage.updateCoins(ctx, tx, fromUserId, -transaction.Amount)
	if err != nil {
		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	err = shopStorage.updateCoins(ctx, tx, toUserId, transaction.Amount)
	if err != nil {
		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	_, err = tx.Exec(ctx, `
		insert into user_transaction(user_from, user_to, money)
		values ($1, $2, $3);
	`, fromUserId, toUserId, transaction.Amount)
	if err != nil {
		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%w (postgres.SendCoin): %w", customErrors.ErrFailedToCommitTx, err)
	}

	return nil
}

func (shopStorage *ShopStorage) BuyItem(ctx context.Context, username string, itemName string) error {
	tx, err := shopStorage.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrFailedToBeginTx, err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			fmt.Printf("%v (postgres.BuyItem): %v", customErrors.ErrFailedToRollbackTx, err)
		}
	}()

	var (
		itemId    int
		itemPrice int
	)
	err = tx.QueryRow(ctx, `
		select id, price
		from product
		where name = $1;
	`, itemName).Scan(&itemId, &itemPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrDoesNotExist, err)
		}

		return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	var (
		userId    int
		userMoney int
	)
	err = tx.QueryRow(ctx, `
		select id, money
		from users
		where name = $1;
	`, username).Scan(&userId, &userMoney)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrDoesNotExist, err)
		}

		return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	if userMoney-itemPrice < 0 {
		return fmt.Errorf("%w (postgres.BuyItem): incorrect amount of coins", customErrors.ErrDataNotValid)
	}

	err = shopStorage.updateCoins(ctx, tx, userId, -itemPrice)
	if err != nil {
		return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	_, err = tx.Exec(ctx, `
		insert into user_product(user_id, product_id)
		values ($1, $2);
	`, userId, itemId)
	if err != nil {
		return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%w (postgres.BuyItem): %w", customErrors.ErrFailedToCommitTx, err)
	}

	return nil
}

func (shopStorage *ShopStorage) updateCoins(ctx context.Context, tx pgx.Tx, userId int, coins int) error {
	_, err := tx.Exec(ctx, `
		update users
		set money = money + $1
		where id = $2;
	`, coins, userId)
	if err != nil {
		return fmt.Errorf("%w (postgres.updateCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return nil
}

func (shopStorage *ShopStorage) getInventory(ctx context.Context, tx pgx.Tx, userId int) ([]domain.Item, error) {
	inventory := make([]domain.Item, 0)
	rows, err := tx.Query(ctx, `
		select p.name, count(*)
		from user_product up, product p
		where up.product_id = p.id and up.user_id = $1
		group by p.id, p.name, bought_at
		order by bought_at desc; 
	`, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w (postgres.getInventory): %w", customErrors.ErrFailedToExecuteQuery, err)
		}
	}
	for rows.Next() {
		var item domain.Item

		err = rows.Scan(&item.Type, &item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("%w (postgres.getInventory): %w", customErrors.ErrFailedToExecuteQuery, err)
		}

		inventory = append(inventory, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w (postgres.getInventory): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return inventory, nil
}

func (shopStorage *ShopStorage) getRecievedCoins(
	ctx context.Context,
	tx pgx.Tx,
	userId int) ([]domain.RecievedCoins, error) {
	recievedCoins := make([]domain.RecievedCoins, 0)
	rows, err := tx.Query(ctx, `
		select u.name, ut.money
		from user_transaction ut, users u
		where ut.user_from = u.id and ut.user_to = $1
		order by sent_at desc; 
	`, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w (postgres.getRecievedCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
		}
	}
	for rows.Next() {
		var coins domain.RecievedCoins

		err = rows.Scan(&coins.From, &coins.Amount)
		if err != nil {
			return nil, fmt.Errorf("%w (postgres.getRecievedCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
		}

		recievedCoins = append(recievedCoins, coins)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w (postgres.getRecievedCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return recievedCoins, nil
}

func (shopStorage *ShopStorage) getSentCoins(ctx context.Context, tx pgx.Tx, userId int) ([]domain.SentCoins, error) {
	sentCoins := make([]domain.SentCoins, 0)
	rows, err := tx.Query(ctx, `
		select u.name, ut.money
		from user_transaction ut, users u
		where ut.user_to = u.id and ut.user_from = $1
		order by sent_at desc; 
	`, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w (postgres.getSentCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
		}
	}
	for rows.Next() {
		var coins domain.SentCoins

		err = rows.Scan(&coins.To, &coins.Amount)
		if err != nil {
			return nil, fmt.Errorf("%w (postgres.getSentCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
		}

		sentCoins = append(sentCoins, coins)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w (postgres.getSentCoins): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return sentCoins, nil
}
