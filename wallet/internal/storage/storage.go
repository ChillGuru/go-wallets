package storage

import (
	"context"
	"errors"
)

type Storage interface {
	CreateWallet(ctx context.Context, name string)
	GetWallet(ctx context.Context, walletID string)
	GetWallets(ctx context.Context)
	UpdateWallet(ctx context.Context)

	BeginTX(ctx context.Context) (Transaction, error)
}

type Wallet struct {
	ID      string
	Name    string
	Balance float64
	Status  string
}

type Transaction interface {
	Commit() error
	Rollback() error
}

// TODO: add more errors
var (
	ErrWalletExists   = errors.New("Wallet already exists")
	ErrWalletNotExist = errors.New("Wallet not exists")
	ErrWalletNotFound = errors.New("Wallet not found")
)
