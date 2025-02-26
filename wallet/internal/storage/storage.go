package storage

import "errors"

//TODO: add more errors
var (
	ErrWalletExists   = errors.New("Wallet already exists")
	ErrWalletNotExist = errors.New("Wallet not exists")
	ErrWalletNotFound = errors.New("Wallet not found")
)
