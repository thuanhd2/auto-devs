package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host            string
	Port            string
	Username        string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnTimeout     time.Duration
}

type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection
func NewConnection(config Config) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool with defaults if not set
	maxOpenConns := config.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 25
	}
	
	maxIdleConns := config.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 5
	}
	
	connMaxLifetime := config.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = 15 * time.Minute
	}
	
	connMaxIdleTime := config.ConnMaxIdleTime
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 5 * time.Minute
	}
	
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// Test the connection with timeout
	testCtx := context.Background()
	if config.ConnTimeout > 0 {
		var cancel context.CancelFunc
		testCtx, cancel = context.WithTimeout(context.Background(), config.ConnTimeout)
		defer cancel()
	}
	
	if err := db.PingContext(testCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.DB.Close()
}

// Ping tests the database connection
func (d *DB) Ping() error {
	return d.DB.Ping()
}

// HealthCheck performs a health check on the database
func (d *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.PingContext(ctx)
}

// BeginTx starts a new transaction with the given options
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.DB.BeginTx(ctx, opts)
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (d *DB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// A panic occurred, rollback and re-panic
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}