package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string) (*DB, error) {
	readOnlyDSN, err := enforceReadOnly(dsn)
	if err != nil {
		return nil, fmt.Errorf("injecting RO mode into database dsn: %s, %w", dsn, err)
	}

	db, err := sql.Open("pgx", readOnlyDSN)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// tune for CLI usage
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// validate connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() {
	d.db.Close()
}

// enforceReadOnly injects postgres runtime RO transactions option
func enforceReadOnly(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parsing dsn: %w", err)
	}

	q := u.Query()
	q.Set("options", "-c default_transaction_read_only=on")

	u.RawQuery = q.Encode()

	return u.String(), nil
}
