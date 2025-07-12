package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/ridopark/jonbu-ohlcv/internal/database"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
	"github.com/ridopark/jonbu-ohlcv/pkg/api/types"
)

// REQ-016: REST API endpoints for historical data access
// REQ-041: Input validation for all endpoints
// REQ-047: Correlation ID for tracing

type OHLCVHandler struct {
	repo   *database.OHLCVRepository
	logger zerolog.Logger
}

// NewOHLCVHandler creates a new OHLCV API handler
func NewOHLCVHandler(repo *database.OHLCVRepository) *OHLCVHandler {
	return &OHLCVHandler{
		repo:   repo,
		logger: logger.NewContextLogger("ohlcv_handler"),
	}
}

// GetOHLCV handles GET /api/v1/ohlcv/{symbol}
func (h *OHLCVHandler) GetOHLCV(w http.ResponseWriter, r *http.Request) {
	// REQ-047: Generate correlation ID for tracing
	correlationID := uuid.New().String()
	reqLogger := logger.NewRequestLogger(correlationID, r.Method, r.URL.Path)

	reqLogger.Info().Msg("Processing OHLCV request")

	// REQ-041: Input validation
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	if err := validateSymbol(symbol); err != nil {
		reqLogger.Error().Err(err).Str("symbol", symbol).Msg("Invalid symbol")
		http.Error(w, "Invalid symbol: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	timeframe := query.Get("timeframe")
	if timeframe == "" {
		timeframe = "1d"
	}

	if err := validateTimeframe(timeframe); err != nil {
		reqLogger.Error().Err(err).Str("timeframe", timeframe).Msg("Invalid timeframe")
		http.Error(w, "Invalid timeframe: "+err.Error(), http.StatusBadRequest)
		return
	}

	limitStr := query.Get("limit")
	limit := 100 // default
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			reqLogger.Error().Str("limit", limitStr).Msg("Invalid limit parameter")
			http.Error(w, "Invalid limit: must be between 1 and 1000", http.StatusBadRequest)
			return
		}
	}

	// Fetch data from repository
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ohlcvs, err := h.repo.GetBySymbol(ctx, symbol, timeframe, limit)
	if err != nil {
		reqLogger.Error().Err(err).Msg("Failed to fetch OHLCV data")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := &types.OHLCVResponse{
		Symbol:    symbol,
		Timeframe: timeframe,
		Count:     len(ohlcvs),
		Data:      ohlcvs,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		reqLogger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	reqLogger.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Int("count", len(ohlcvs)).
		Int("limit", limit).
		Msg("OHLCV request completed successfully")
}

// GetOHLCVHistory handles GET /api/v1/ohlcv/{symbol}/history
func (h *OHLCVHandler) GetOHLCVHistory(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	reqLogger := logger.NewRequestLogger(correlationID, r.Method, r.URL.Path)

	reqLogger.Info().Msg("Processing OHLCV history request")

	// REQ-041: Input validation
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	if err := validateSymbol(symbol); err != nil {
		reqLogger.Error().Err(err).Str("symbol", symbol).Msg("Invalid symbol")
		http.Error(w, "Invalid symbol: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	timeframe := query.Get("timeframe")
	if timeframe == "" {
		timeframe = "1d"
	}

	if err := validateTimeframe(timeframe); err != nil {
		reqLogger.Error().Err(err).Str("timeframe", timeframe).Msg("Invalid timeframe")
		http.Error(w, "Invalid timeframe: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse time range
	startStr := query.Get("start")
	endStr := query.Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			reqLogger.Error().Err(err).Str("start", startStr).Msg("Invalid start date format")
			http.Error(w, "Invalid start date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		// Default to 30 days ago
		start = time.Now().AddDate(0, 0, -30)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			reqLogger.Error().Err(err).Str("end", endStr).Msg("Invalid end date format")
			http.Error(w, "Invalid end date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		end = time.Now()
	}

	// Validate date range
	if start.After(end) {
		reqLogger.Error().Time("start", start).Time("end", end).Msg("Invalid date range")
		http.Error(w, "Start date must be before end date", http.StatusBadRequest)
		return
	}

	// Parse limit
	limitStr := query.Get("limit")
	limit := 1000 // default for history
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 10000 {
			reqLogger.Error().Str("limit", limitStr).Msg("Invalid limit parameter")
			http.Error(w, "Invalid limit: must be between 1 and 10000", http.StatusBadRequest)
			return
		}
	}

	// Fetch historical data
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	ohlcvs, err := h.repo.GetHistory(ctx, symbol, timeframe, start, end, limit)
	if err != nil {
		reqLogger.Error().Err(err).Msg("Failed to fetch OHLCV history")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := &types.OHLCVHistoryResponse{
		Symbol:    symbol,
		Timeframe: timeframe,
		Start:     start,
		End:       end,
		Count:     len(ohlcvs),
		Data:      ohlcvs,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		reqLogger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	reqLogger.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Time("start", start).
		Time("end", end).
		Int("count", len(ohlcvs)).
		Msg("OHLCV history request completed successfully")
}

// GetLatestOHLCV handles GET /api/v1/ohlcv/{symbol}/latest
func (h *OHLCVHandler) GetLatestOHLCV(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	reqLogger := logger.NewRequestLogger(correlationID, r.Method, r.URL.Path)

	reqLogger.Info().Msg("Processing latest OHLCV request")

	// REQ-041: Input validation
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	if err := validateSymbol(symbol); err != nil {
		reqLogger.Error().Err(err).Str("symbol", symbol).Msg("Invalid symbol")
		http.Error(w, "Invalid symbol: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse timeframe
	query := r.URL.Query()
	timeframe := query.Get("timeframe")
	if timeframe == "" {
		timeframe = "1d"
	}

	if err := validateTimeframe(timeframe); err != nil {
		reqLogger.Error().Err(err).Str("timeframe", timeframe).Msg("Invalid timeframe")
		http.Error(w, "Invalid timeframe: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch latest data
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ohlcv, err := h.repo.GetLatest(ctx, symbol, timeframe)
	if err != nil {
		reqLogger.Error().Err(err).Msg("Failed to fetch latest OHLCV")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if ohlcv == nil {
		reqLogger.Info().Str("symbol", symbol).Str("timeframe", timeframe).Msg("No data found")
		http.Error(w, "No data found for symbol", http.StatusNotFound)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(ohlcv); err != nil {
		reqLogger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	reqLogger.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Time("timestamp", ohlcv.Timestamp).
		Msg("Latest OHLCV request completed successfully")
}

// REQ-041: Input validation functions
func validateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if len(symbol) > 10 {
		return fmt.Errorf("symbol too long: maximum 10 characters")
	}

	for _, char := range symbol {
		if char < 'A' || char > 'Z' {
			return fmt.Errorf("symbol must contain only uppercase letters")
		}
	}

	return nil
}

func validateTimeframe(timeframe string) error {
	validTimeframes := map[string]bool{
		"1m":  true,
		"5m":  true,
		"15m": true,
		"1h":  true,
		"4h":  true,
		"1d":  true,
	}

	if !validTimeframes[timeframe] {
		return fmt.Errorf("invalid timeframe: must be one of 1m, 5m, 15m, 1h, 4h, 1d")
	}

	return nil
}
