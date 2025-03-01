package storage

import "errors"

type Wallet struct {
	ID      string
	Name    string
	Balance float64
	Status  string
}

//TODO: add more errors
var (
	ErrWalletExists   = errors.New("Wallet already exists")
	ErrWalletNotExist = errors.New("Wallet not exists")
	ErrWalletNotFound = errors.New("Wallet not found")
)
