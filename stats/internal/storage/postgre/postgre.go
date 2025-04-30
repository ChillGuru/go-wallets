package postgre

import (
	"context"
	"database/sql"
	"fmt"
	"stats/internal/storage"
	"stats/internal/utils/random"
	"time"

	_ "github.com/lib/pq"
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
		CREATE TABLE IF NOT EXISTS stats(
			id TEXT PRIMARY KEY,
			total INTEGER NOT NULL,
			active INTEGER NOT NULL,
			inactive INTEGER NOT NULL,
			deposited FLOAT NOT NULL,
			withdrawn FLOAT NOT NULL,
			transfered FLOAT NOT NULL,
			operation VARCHAR(24) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	// Проверяем, пустая ли таблица
	var exists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM stats LIMIT 1)
	`).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("check table emptiness error: %w", err)
	}

	// Если таблица пустая - вставляем дефолтную запись
	if !exists {
		statsID := random.NewRandomString(ID_LENGTH)
		query := `
			INSERT INTO stats (
				id,
            	total, 
            	active, 
            	inactive, 
            	deposited, 
            	withdrawn, 
            	transfered,
				operation
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
		`
		_, err := db.ExecContext(
			ctx,
			query,
			statsID,
			0,
			0,
			0,
			0.0,
			0.0,
			0.0,
			"STATS BEGIN",
		)

		if err != nil {
			return nil, fmt.Errorf("initial insert error: %w", err)
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
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

func (t *PostgreTx) GetStats(ctx context.Context) (*storage.Stats, error) {
	const fn = "postgre.GetStats"

	stmt, err := t.tx.Prepare(`
		SELECT 
			id, 
			total, 
			active, 
			inactive, 
			deposited, 
			withdrawn, 
			transfered,
			created_at 
			FROM stats 
			ORDER BY created_at DESC 
			LIMIT 1
		`)
	if err != nil {
		return nil, fmt.Errorf("%s failed to prepare query for get stats: %w", fn, err)
	}

	defer stmt.Close()

	var stats storage.Stats

	var createdAt time.Time

	err = stmt.QueryRowContext(ctx).Scan(
		&stats.ID,
		&stats.Total,
		&stats.Active,
		&stats.Inactive,
		&stats.Deposited,
		&stats.Withdrawn,
		&stats.Transfered,
		&createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed to to get stats: %w", fn, err)
	}

	return &stats, nil
}

func (t *PostgreTx) UpdateStats(ctx context.Context, operation string, amount ...float64) error {
	const fn = "postgre.UpdateStats"

	currentStats, err := t.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	newStats := *currentStats
	newStats.ID = random.NewRandomString(ID_LENGTH)

	switch operation {
	case storage.OpCreate:
		newStats.Total++
		newStats.Active++
	case storage.OpDelete:
		newStats.Inactive++
		newStats.Active--
	case storage.OpDeposit, storage.OpWithdraw, storage.OpTransfer:
		if len(amount) == 0 || amount[0] <= 0 {
			return fmt.Errorf("%s: amount is required for %s operation", fn, operation)
		}

		switch operation {
		case storage.OpDeposit:
			newStats.Deposited += amount[0]
		case storage.OpWithdraw:
			newStats.Withdrawn += amount[0]
		case storage.OpTransfer:
			newStats.Transfered += amount[0]
		}
	default:
		return fmt.Errorf("%s: unknown operation %s", fn, operation)
	}

	stmt, err := t.tx.Prepare(`
		INSERT INTO stats (
		id, total, active, inactive, 
		deposited, withdrawn, transfered,
		operation
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		newStats.ID,
		newStats.Total,
		newStats.Active,
		newStats.Inactive,
		newStats.Deposited,
		newStats.Withdrawn,
		newStats.Transfered,
		operation,
	)

	if err != nil {
		return fmt.Errorf("%s: failed to insert new stats: %w", fn, err)
	}

	return nil
}
