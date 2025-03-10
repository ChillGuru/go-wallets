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
	ID      string
	Name    string
	Balance float64
	Status  string
}

// TODO: add more errors
var (
	ErrWalletExists   = errors.New("Wallet already exists")
	ErrWalletNotExist = errors.New("Wallet not exists")
	ErrWalletNotFound = errors.New("Wallet not found")
)
