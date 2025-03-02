package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"wallet/internal/lib/random"
	"wallet/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

const ID_LENGTH = 16

func New(path string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	//open bd
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	//убрать
	if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	//check bd with ping
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS wallet(
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			balance FLOAT DEFAULT 0.0,
			status TEXT DEFAULT 'active');
		CREATE INDEX IF NOT EXISTS idx_name ON wallet(name);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateWallet(ctx context.Context, name string) (int64, error) {
	const fn = "sqlite.CreateWallet"

	stmt, err := s.db.Prepare(`INSERT INTO wallet(id, name) VALUES(?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for creating wallet: %w", fn, err)
	}

	defer stmt.Close()

	walletID := random.NewRandomString(ID_LENGTH)

	res, err := stmt.ExecContext(ctx, walletID, name)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrWalletExists)
		}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetWallet(ctx context.Context, walletID string) (*storage.Wallet, error) {
	const fn = "sqlite.GetWallet"

	stmt, err := s.db.Prepare(`SELECT id, name, balance, status FROM wallet WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallet: %w", fn, err)
	}

	defer stmt.Close()

	var wallet storage.Wallet

	err = stmt.QueryRowContext(ctx, walletID).Scan(&wallet.ID, &wallet.Name, &wallet.Balance, &wallet.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s failed to get wallet: %w", fn, storage.ErrWalletNotExist)
	}

	return &wallet, nil
}

func (s *Storage) GetWallets(ctx context.Context) ([]storage.Wallet, error) {
	const fn = "sqlite.GetWallets"

	stmt, err := s.db.Prepare(`SELECT id, name, balance, status FROM wallet`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallets: %w", fn, err)
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)

	var wallets []storage.Wallet

	for rows.Next() {
		var wallet storage.Wallet

		if err := rows.Scan(&wallet.ID, &wallet.Name, &wallet.Balance, &wallet.Status); err != nil {
			return nil, fmt.Errorf("%s failed to scan row: %w", fn, err)
		}

		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return wallets, nil
}

func (s *Storage) UpdateWallet(ctx context.Context, updatedWallet *storage.Wallet) (int64, error) {
	const fn = "sqlite.UpdateWallet"

	stmt, err := s.db.Prepare(`UPDATE wallet SET name = ?, balance = ?, status = ? WHERE id = ?`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for update wallet: %w", fn, err)
	}

	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, updatedWallet.Name, updatedWallet.Balance, updatedWallet.Status, updatedWallet.ID)
	if err != nil {
		return 0, fmt.Errorf("%s failed to update wallet: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) DeactivateWallet(ctx context.Context, walletID string) (int64, error) {
	const fn = "sqlite.DeactivateWallet"

	stmt, err := s.db.Prepare(`UPDATE wallet SET status = "inactive" WHERE id = ?`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for deactivate wallet: %w", fn, err)
	}

	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("%s failed to deactivate wallet: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) BeginTx(ctx context.Context) (storage.Transaction, error) {
	const fn = "sqlite.BeginTx"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s failed to begin transaction: %w", fn, err)
	}

	return &SQLiteTx{tx: tx}, nil
}

type SQLiteTx struct {
	tx *sql.Tx
}

func (t *SQLiteTx) Commit() error {
	return t.tx.Commit()
}

func (t *SQLiteTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *SQLiteTx) GetWallet(ctx context.Context, walletID string) (*storage.Wallet, error) {
	const fn = "sqlite.GetWallet"

	stmt, err := t.tx.Prepare(`SELECT id, name, balance, status FROM wallet WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallet: %w", fn, err)
	}

	defer stmt.Close()

	var wallet storage.Wallet

	err = stmt.QueryRowContext(ctx, walletID).Scan(&wallet.ID, &wallet.Name, &wallet.Balance, &wallet.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s failed to get wallet: %w", fn, storage.ErrWalletNotExist)
	}

	return &wallet, nil
}

func (t *SQLiteTx) UpdateWallet(ctx context.Context, updatedWallet *storage.Wallet) (int64, error) {
	const fn = "sqlite.UpdateWallet"

	stmt, err := t.tx.Prepare(`UPDATE wallet SET name = ?, balance = ?, status = ? WHERE id = ?`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for update wallet: %w", fn, err)
	}

	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, updatedWallet.Name, updatedWallet.Balance, updatedWallet.Status, updatedWallet.ID)
	if err != nil {
		return 0, fmt.Errorf("%s failed to update wallet: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id: %w", fn, err)
	}

	return id, nil
}
