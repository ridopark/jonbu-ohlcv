package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/ridopark/jonbu-ohlcv/internal/logger"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-013: Use prepared statements for optimal performance
// REQ-015: Proper transaction handling for batch operations

type OHLCVRepository struct {
	db     *DB
	logger zerolog.Logger

	// Prepared statements for performance
	insertStmt         *sql.Stmt
	selectBySymbolStmt *sql.Stmt
	selectHistoryStmt  *sql.Stmt
	selectLatestStmt   *sql.Stmt
}

// NewOHLCVRepository creates a new OHLCV repository with prepared statements
func NewOHLCVRepository(db *DB) (*OHLCVRepository, error) {
	logger := logger.NewContextLogger("ohlcv_repository")

	repo := &OHLCVRepository{
		db:     db,
		logger: logger,
	}

	// REQ-013: Prepare statements for optimal performance
	if err := repo.prepareStatements(); err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return repo, nil
}

// Close closes all prepared statements
func (r *OHLCVRepository) Close() error {
	statements := []*sql.Stmt{
		r.insertStmt,
		r.selectBySymbolStmt,
		r.selectHistoryStmt,
		r.selectLatestStmt,
	}

	for _, stmt := range statements {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				r.logger.Error().Err(err).Msg("Failed to close prepared statement")
			}
		}
	}

	return nil
}

// Insert stores a single OHLCV record
func (r *OHLCVRepository) Insert(ctx context.Context, ohlcv *models.OHLCV) error {
	start := time.Now()
	defer func() {
		logger.LogPerformance(r.logger, "insert_ohlcv", start, true)
	}()

	ohlcv.CreatedAt = time.Now()
	ohlcv.UpdatedAt = time.Now()

	err := r.insertStmt.QueryRowContext(
		ctx,
		ohlcv.Symbol,
		ohlcv.Timestamp,
		ohlcv.Open,
		ohlcv.High,
		ohlcv.Low,
		ohlcv.Close,
		ohlcv.Volume,
		ohlcv.Timeframe,
		ohlcv.CreatedAt,
		ohlcv.UpdatedAt,
	).Scan(&ohlcv.ID)

	if err != nil {
		logger.LogError(r.logger, err, "Failed to insert OHLCV record", map[string]interface{}{
			"symbol":    ohlcv.Symbol,
			"timestamp": ohlcv.Timestamp,
			"timeframe": ohlcv.Timeframe,
		})
		return fmt.Errorf("failed to insert OHLCV: %w", err)
	}

	r.logger.Debug().
		Str("symbol", ohlcv.Symbol).
		Time("timestamp", ohlcv.Timestamp).
		Str("timeframe", ohlcv.Timeframe).
		Int64("id", ohlcv.ID).
		Msg("OHLCV record inserted")

	return nil
}

// REQ-015: Batch insert with transaction support
func (r *OHLCVRepository) InsertBatch(ctx context.Context, ohlcvs []*models.OHLCV) error {
	if len(ohlcvs) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		logger.LogPerformance(r.logger, "insert_batch_ohlcv", start, true)
	}()

	return r.db.ExecuteInTransaction(ctx, func(tx *sql.Tx) error {
		stmt := tx.Stmt(r.insertStmt)
		defer stmt.Close()

		for _, ohlcv := range ohlcvs {
			ohlcv.CreatedAt = time.Now()
			ohlcv.UpdatedAt = time.Now()

			err := stmt.QueryRowContext(
				ctx,
				ohlcv.Symbol,
				ohlcv.Timestamp,
				ohlcv.Open,
				ohlcv.High,
				ohlcv.Low,
				ohlcv.Close,
				ohlcv.Volume,
				ohlcv.Timeframe,
				ohlcv.CreatedAt,
				ohlcv.UpdatedAt,
			).Scan(&ohlcv.ID)

			if err != nil {
				return fmt.Errorf("failed to insert OHLCV batch record: %w", err)
			}
		}

		r.logger.Info().
			Int("count", len(ohlcvs)).
			Msg("OHLCV batch inserted successfully")

		return nil
	})
}

// GetBySymbol retrieves OHLCV records for a specific symbol
func (r *OHLCVRepository) GetBySymbol(ctx context.Context, symbol, timeframe string, limit int) ([]*models.OHLCV, error) {
	start := time.Now()
	defer func() {
		logger.LogPerformance(r.logger, "get_by_symbol", start, true)
	}()

	rows, err := r.selectBySymbolStmt.QueryContext(ctx, symbol, timeframe, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query OHLCV by symbol: %w", err)
	}
	defer rows.Close()

	var result []*models.OHLCV
	for rows.Next() {
		ohlcv := &models.OHLCV{}
		err := rows.Scan(
			&ohlcv.ID,
			&ohlcv.Symbol,
			&ohlcv.Timestamp,
			&ohlcv.Open,
			&ohlcv.High,
			&ohlcv.Low,
			&ohlcv.Close,
			&ohlcv.Volume,
			&ohlcv.Timeframe,
			&ohlcv.CreatedAt,
			&ohlcv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OHLCV row: %w", err)
		}
		result = append(result, ohlcv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating OHLCV rows: %w", err)
	}

	return result, nil
}

// GetHistory retrieves historical OHLCV data with time range filtering
func (r *OHLCVRepository) GetHistory(ctx context.Context, symbol, timeframe string, start, end time.Time, limit int) ([]*models.OHLCV, error) {
	start_time := time.Now()
	defer func() {
		logger.LogPerformance(r.logger, "get_history", start_time, true)
	}()

	rows, err := r.selectHistoryStmt.QueryContext(ctx, symbol, timeframe, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query OHLCV history: %w", err)
	}
	defer rows.Close()

	var result []*models.OHLCV
	for rows.Next() {
		ohlcv := &models.OHLCV{}
		err := rows.Scan(
			&ohlcv.ID,
			&ohlcv.Symbol,
			&ohlcv.Timestamp,
			&ohlcv.Open,
			&ohlcv.High,
			&ohlcv.Low,
			&ohlcv.Close,
			&ohlcv.Volume,
			&ohlcv.Timeframe,
			&ohlcv.CreatedAt,
			&ohlcv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OHLCV history row: %w", err)
		}
		result = append(result, ohlcv)
	}

	return result, nil
}

// GetLatest retrieves the most recent OHLCV record for a symbol
func (r *OHLCVRepository) GetLatest(ctx context.Context, symbol, timeframe string) (*models.OHLCV, error) {
	start := time.Now()
	defer func() {
		logger.LogPerformance(r.logger, "get_latest", start, true)
	}()

	ohlcv := &models.OHLCV{}
	err := r.selectLatestStmt.QueryRowContext(ctx, symbol, timeframe).Scan(
		&ohlcv.ID,
		&ohlcv.Symbol,
		&ohlcv.Timestamp,
		&ohlcv.Open,
		&ohlcv.High,
		&ohlcv.Low,
		&ohlcv.Close,
		&ohlcv.Volume,
		&ohlcv.Timeframe,
		&ohlcv.CreatedAt,
		&ohlcv.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest OHLCV: %w", err)
	}

	return ohlcv, nil
}

// prepareStatements prepares all SQL statements for optimal performance
func (r *OHLCVRepository) prepareStatements() error {
	var err error

	// Insert statement with RETURNING clause
	insertSQL := `
		INSERT INTO ohlcv (symbol, timestamp, open, high, low, close, volume, timeframe, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	r.insertStmt, err = r.db.conn.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// Select by symbol statement
	selectBySymbolSQL := `
		SELECT id, symbol, timestamp, open, high, low, close, volume, timeframe, created_at, updated_at
		FROM ohlcv
		WHERE symbol = $1 AND timeframe = $2
		ORDER BY timestamp DESC
		LIMIT $3`

	r.selectBySymbolStmt, err = r.db.conn.Prepare(selectBySymbolSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare select by symbol statement: %w", err)
	}

	// Select history statement
	selectHistorySQL := `
		SELECT id, symbol, timestamp, open, high, low, close, volume, timeframe, created_at, updated_at
		FROM ohlcv
		WHERE symbol = $1 AND timeframe = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp ASC
		LIMIT $5`

	r.selectHistoryStmt, err = r.db.conn.Prepare(selectHistorySQL)
	if err != nil {
		return fmt.Errorf("failed to prepare select history statement: %w", err)
	}

	// Select latest statement
	selectLatestSQL := `
		SELECT id, symbol, timestamp, open, high, low, close, volume, timeframe, created_at, updated_at
		FROM ohlcv
		WHERE symbol = $1 AND timeframe = $2
		ORDER BY timestamp DESC
		LIMIT 1`

	r.selectLatestStmt, err = r.db.conn.Prepare(selectLatestSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare select latest statement: %w", err)
	}

	r.logger.Info().Msg("All prepared statements created successfully")
	return nil
}
