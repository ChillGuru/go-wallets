package service

import (
	"context"
	"fmt"
	"wallet/internal/storage"
)

type WalletService struct {
	storage storage.Storage
}

func New(storage storage.Storage) *WalletService {
	return &WalletService{storage: storage}
}

func (w *WalletService) Deposit(ctx context.Context, walletID string, amount float64) (int64, error) {
	const fn = "WalletService.Deposit"
	if amount <= 0 {
		return 0, fmt.Errorf("%s: Amount must be positive", fn)
	}

	tx, err := w.storage.BeginTx(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	wallet, err := tx.GetWallet(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	wallet.Balance += amount

	id, err := tx.UpdateWallet(ctx, wallet)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (w *WalletService) Withdraw(ctx context.Context, walletID string, amount float64) (int64, error) {
	const fn = "WalletService.Withdraw"
	if amount <= 0 {
		return 0, fmt.Errorf("%s: Amount must be positive", fn)
	}

	tx, err := w.storage.BeginTx(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	wallet, err := tx.GetWallet(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if wallet.Balance < amount {
		return 0, fmt.Errorf("%s: Insufficient funds", fn)
	}

	wallet.Balance -= amount

	id, err := tx.UpdateWallet(ctx, wallet)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (w *WalletService) Transfer(ctx context.Context, walletID string, amount float64, transferTo string) (int64, int64, error) {
	const fn = "WalletService.Transfer"
	if amount <= 0 {
		return 0, 0, fmt.Errorf("%s: Amount must be positive", fn)
	}

	tx, err := w.storage.BeginTx(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}

	fromWallet, err := tx.GetWallet(ctx, walletID)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}

	toWallet, err := tx.GetWallet(ctx, transferTo)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}

	if fromWallet.Balance < amount {
		return 0, 0, fmt.Errorf("%s: Insufficient funds", fn)
	}

	fromWallet.Balance -= amount
	toWallet.Balance += amount

	id, err := tx.UpdateWallet(ctx, fromWallet)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}

	recipientID, err := tx.UpdateWallet(ctx, toWallet)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, recipientID, nil
}
