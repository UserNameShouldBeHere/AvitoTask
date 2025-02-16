package domain

import (
	"fmt"

	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

type Transaction struct {
	From   string
	To     string
	Amount int
}

func (transaction *Transaction) Validate() error {
	if len(transaction.From) < 3 ||
		len(transaction.From) >= 150 {
		return fmt.Errorf("%w (Validate): incorrect name length", customErrors.ErrDataNotValid)
	}

	if len(transaction.To) < 3 ||
		len(transaction.To) >= 150 {
		return fmt.Errorf("%w (Validate): incorrect name length", customErrors.ErrDataNotValid)
	}

	if transaction.Amount < 0 {
		return fmt.Errorf("%w (Validate): incorrect amount of coins", customErrors.ErrDataNotValid)
	}

	return nil
}

type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type RecievedCoins struct {
	From   string `json:"fromUser"`
	Amount int    `json:"amount"`
}

type SentCoins struct {
	To     string `json:"toUser"`
	Amount int    `json:"amount"`
}

type SentRecievedHistory struct {
	Recieved []RecievedCoins `json:"recieved"`
	Sent     []SentCoins     `json:"sent"`
}

type InventoryInfo struct {
	Coins       int                 `json:"coing"`
	Inventory   []Item              `json:"inventory"`
	CoinHistory SentRecievedHistory `json:"coinHistory"`
}
