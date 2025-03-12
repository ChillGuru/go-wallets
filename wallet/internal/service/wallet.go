package service

import (
	"context"
	"errors"
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
		return 0, err
	}

	wallet, err := tx.GetWallet(ctx, walletID)
	if err != nil {
		return 0, err
	}
	if wallet.Status == "inactive" {
		return 0, storage.ErrWalletNotFound
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
	if wallet.Status == "inactive" {
		return 0, storage.ErrWalletNotFound
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
	if fromWallet.Status == "inactive" {
		return 0, 0, storage.ErrWalletNotFound
	}

	toWallet, err := tx.GetWallet(ctx, transferTo)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: %w", fn, err)
	}
	if toWallet.Status == "inactive" {
		return 0, 0, storage.ErrWalletNotFound
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

func (w *WalletService) UpdateName(ctx context.Context, walletID, name string) (int64, error) {
	const fn = "WalletService.UpdateName"
	if len(name) <= 1 {
		return 0, fmt.Errorf("%s: The name length must be more than 1 character", fn)
	}

	tx, err := w.storage.BeginTx(ctx)
	if err != nil {
		return 0, err
	}

	wallet, err := tx.GetWallet(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	if wallet.Status == "inactive" {
		return 0, storage.ErrWalletNotFound
	}

	wallet.Name = name

	id, err := tx.UpdateWallet(ctx, wallet)
	if errors.Is(err, storage.ErrWalletNotExist) {
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (w *WalletService) CreateWallet(ctx context.Context, name string) (*storage.Wallet, error) {
	const fn = "WalletService.CreateWallet"
	if len(name) <= 1 {
		return nil, fmt.Errorf("%s: The name length must be more than 1 character", fn)
	}

	walletID, err := w.storage.CreateWallet(ctx, name)
	if errors.Is(err, storage.ErrWalletExists) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &storage.Wallet{ID: walletID, Name: name, Status: "active"}, nil
}

func (w *WalletService) GetWallet(ctx context.Context, walletID string) (*storage.Wallet, error) {
	const fn = "WalletService.GetWallet"

	var wallet *storage.Wallet

	wallet, err := w.storage.GetWallet(ctx, walletID)
	if errors.Is(err, storage.ErrWalletNotExist) {
		return nil, err
	}
	if wallet.Status == "inactive" {
		return nil, storage.ErrWalletNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return wallet, nil
}

func (w *WalletService) GetWallets(ctx context.Context) ([]storage.Wallet, error) {
	const fn = "WalletService.GetWallets"

	wallets, err := w.storage.GetWallets(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return wallets, nil
}

func (w *WalletService) DeactivateWallet(ctx context.Context, walletID string) (int64, error) {
	const fn = "WalletService.GetWallets"

	id, err := w.storage.DeactivateWallet(ctx, walletID)
	if errors.Is(err, storage.ErrWalletNotExist) {
		return 0, err
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}
