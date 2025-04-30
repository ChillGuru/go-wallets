package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"wallet/internal/storage"
	"wallet/internal/utils/random"

	_ "github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

const ID_LENGTH = 16

func New(dbhost string, dbport int) (*Storage, error) {
	const fn = "storage.postgre.New"
	const (
		user     = "wallets_admin"
		password = "admin"
		dbname   = "wallets_db"
	)

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbhost, dbport, user, password, dbname,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS wallet(
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			balance FLOAT DEFAULT 0.0,
			status TEXT DEFAULT 'active');
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

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateWallet(ctx context.Context, name string) (string, error) {
	const fn = "postgre.CreateWallet"

	stmt, err := s.db.Prepare(`INSERT INTO wallet(id, name) VALUES($1, $2)`)
	if err != nil {
		return "", fmt.Errorf("%s failed to prepare query for creating wallet: %w", fn, err)
	}

	defer stmt.Close()

	walletID := random.NewRandomString(ID_LENGTH)

	_, err = stmt.ExecContext(ctx, walletID, name)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return "", fmt.Errorf("%s: %w", fn, storage.ErrWalletExists)
		}
	}

	return walletID, nil
}

func (s *Storage) GetWallet(ctx context.Context, walletID string) (*storage.Wallet, error) {
	const fn = "postgre.GetWallet"

	stmt, err := s.db.Prepare(`SELECT id, name, balance, status FROM wallet WHERE id = $1`)
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
	const fn = "postgre.GetWallets"

	stmt, err := s.db.Prepare(`SELECT id, name, balance, status FROM wallet`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallets: %w", fn, err)
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

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
	const fn = "postgre.UpdateWallet"

	stmt, err := s.db.Prepare(`UPDATE wallet SET name = $1, balance = $2, status = $3 WHERE id = $4`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for update wallet: %w", fn, err)
	}

	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, updatedWallet.Name, updatedWallet.Balance, updatedWallet.Status, updatedWallet.ID)
	if err != nil {
		return 0, fmt.Errorf("%s failed to update wallet: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get affected rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return 0, storage.ErrWalletNotExist
	}

	return rowsAffected, nil
}

func (s *Storage) DeactivateWallet(ctx context.Context, walletID string) (int64, error) {
	const fn = "postgre.DeactivateWallet"

	stmt, err := s.db.Prepare(`UPDATE wallet SET status = 'inactive' WHERE id = $1`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for deactivate wallet: %w", fn, err)
	}

	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("%s failed to deactivate wallet: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get affected rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return 0, storage.ErrWalletNotExist
	}

	return rowsAffected, nil
}

func (s *Storage) BeginTx(ctx context.Context) (storage.Transaction, error) {
	const fn = "postgre.BeginTx"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s failed to begin transaction: %w", fn, err)
	}

	return &PostgreTx{tx: tx}, nil
}

type PostgreTx struct {
	tx *sql.Tx
}

func (t *PostgreTx) Commit() error {
	return t.tx.Commit()
}

func (t *PostgreTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *PostgreTx) GetWallet(ctx context.Context, walletID string) (*storage.Wallet, error) {
	const fn = "postgre.GetWallet"

	stmt, err := t.tx.Prepare(`SELECT id, name, balance, status FROM wallet WHERE id = $1`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallet: %w", fn, err)
	}

	defer stmt.Close()

	var wallet storage.Wallet

	err = stmt.QueryRowContext(ctx, walletID).Scan(&wallet.ID, &wallet.Name, &wallet.Balance, &wallet.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrWalletNotExist
	}

	return &wallet, nil
}

func (t *PostgreTx) UpdateWallet(ctx context.Context, updatedWallet *storage.Wallet) (int64, error) {
	const fn = "postgre.UpdateWallet"

	stmt, err := t.tx.Prepare(`UPDATE wallet SET name = $1, balance = $2, status = $3 WHERE id = $4`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for update wallet: %w", fn, err)
	}

	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, updatedWallet.Name, updatedWallet.Balance, updatedWallet.Status, updatedWallet.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, storage.ErrWalletNotExist
	}
	if err != nil {
		return 0, fmt.Errorf("%s failed to update wallet: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get affected rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return 0, storage.ErrWalletNotExist
	}

	return rowsAffected, nil
}
