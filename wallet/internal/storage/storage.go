package storage

import (
	"context"
	"errors"
)

type Storage interface {
	CreateWallet(ctx context.Context, name string) (string, error)
	GetWallet(ctx context.Context, walletID string) (*Wallet, error)
	GetWallets(ctx context.Context) ([]Wallet, error)
	UpdateWallet(ctx context.Context, updatedWallet *Wallet) (int64, error)
	DeactivateWallet(ctx context.Context, walletID string) (int64, error)
	//Транзакции
	BeginTx(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
	GetWallet(ctx context.Context, walletID string) (*Wallet, error)
	UpdateWallet(ctx context.Context, updatedWallet *Wallet) (int64, error)
}

type Wallet struct {
	ID      string  `json:"id"`
	Name    string  `json:"name,omitempty"`
	Balance float64 `json:"balance,omitempty"`
	Status  string  `json:"status,omitempty"`
}

// TODO: add more errors
var (
	ErrWalletExists   = errors.New("wallet already exists")
	ErrWalletNotExist = errors.New("wallet not exists")
	ErrWalletNotFound = errors.New("wallet not found")
)
