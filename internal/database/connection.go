package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog"

	"github.com/ridopark/jonbu-ohlcv/internal/config"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
)

// REQ-011: PostgreSQL database storage
// REQ-014: Connection pooling management
// REQ-015: Proper transaction handling

type DB struct {
	conn   *sql.DB
	logger zerolog.Logger
}

// NewConnection creates a new database connection with pooling
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	logger := logger.NewContextLogger("database")

	// REQ-014: Build connection string with pooling parameters
	connStr := buildConnectionString(cfg)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// REQ-014: Configure connection pool
	conn.SetMaxOpenConns(cfg.MaxConnections)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Int("max_connections", cfg.MaxConnections).
		Msg("Database connection established")

	return &DB{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying sql.DB connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// Ping checks if the database is reachable
func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// REQ-015: Transaction support for batch operations
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, opts)
}

// ExecuteInTransaction executes a function within a database transaction
func (db *DB) ExecuteInTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				db.logger.Error().
					Err(rbErr).
					Msg("Failed to rollback transaction")
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				db.logger.Error().
					Err(commitErr).
					Msg("Failed to commit transaction")
				err = commitErr
			}
		}
	}()

	err = fn(tx)
	return err
}

// HealthCheck performs a comprehensive database health check
func (db *DB) HealthCheck(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})

	// Check basic connectivity
	if err := db.Ping(ctx); err != nil {
		result["status"] = "unhealthy"
		result["error"] = err.Error()
		return result
	}

	// Get connection stats
	stats := db.conn.Stats()
	result["status"] = "healthy"
	result["open_connections"] = stats.OpenConnections
	result["in_use"] = stats.InUse
	result["idle"] = stats.Idle
	result["wait_count"] = stats.WaitCount
	result["wait_duration"] = stats.WaitDuration.String()
	result["max_idle_closed"] = stats.MaxIdleClosed
	result["max_idle_time_closed"] = stats.MaxIdleTimeClosed
	result["max_lifetime_closed"] = stats.MaxLifetimeClosed

	return result
}

// buildConnectionString constructs the PostgreSQL connection string
func buildConnectionString(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)
}

// IsConnectionError checks if an error is a connection-related error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for PostgreSQL connection errors
	if pqErr, ok := err.(*pq.Error); ok {
		// Connection errors typically have these codes
		switch pqErr.Code {
		case "08000", "08003", "08006", "08001", "08004":
			return true
		}
	}

	// Check for context timeout or cancellation
	if err == context.DeadlineExceeded || err == context.Canceled {
		return true
	}

	return false
}
