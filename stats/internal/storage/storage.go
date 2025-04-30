package storage

import (
	"context"
)

const (
	OpCreate   = "create"
	OpDelete   = "delete"
	OpDeposit  = "deposit"
	OpWithdraw = "withdraw"
	OpTransfer = "transfer"
)

type Storage interface {
	Close() error
	//Транзакции
	BeginTx(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
	UpdateStats(ctx context.Context, operation string, amount ...float64) error
	GetStats(ctx context.Context) (*Stats, error)
}

type Stats struct {
	ID         string  `json:"id"`
	Total      int     `json:"total"`
	Active     int     `json:"active"`
	Inactive   int     `json:"inactive"`
	Deposited  float64 `json:"deposited"`
	Withdrawn  float64 `json:"withdrawn"`
	Transfered float64 `json:"transfered"`
}
