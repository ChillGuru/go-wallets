package sqlite

import (
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

func (s Storage) CreateWallet(name string) (int64, error) {
	const fn = "sqlite.CreateWallet"

	stmt, err := s.db.Prepare(`INSERT INTO wallet(id, name) VALUES(?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s failed to prepare query for creating wallet: %w", fn, err)
	}

	defer stmt.Close()

	walletID := random.NewRandomString(ID_LENGTH)

	res, err := stmt.Exec(walletID, name)
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

func (s Storage) GetWallet(walletID string) (*storage.Wallet, error) {
	const fn = "sqlite.GetWallet"

	stmt, err := s.db.Prepare(`SELECT id, name, balance, status FROM wallet WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallet: %w", fn, err)
	}

	defer stmt.Close()

	var wallet storage.Wallet

	err = stmt.QueryRow(walletID).Scan(&wallet.ID, &wallet.Name, &wallet.Balance, &wallet.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s failed to get wallet: %w", fn, storage.ErrWalletNotExist)
	}

	return &wallet, nil
}

func (s Storage) GetWallets() ([]storage.Wallet, error) {
	const fn = "sqlite.GetWallets"

	stmt, err := s.db.Prepare(`SELECT id, name, balance, status FROM wallet`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get wallets: %w", fn, err)
	}

	defer stmt.Close()

	rows, err := stmt.Query()

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
