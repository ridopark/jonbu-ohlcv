package types

import (
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-016: API request/response types for REST endpoints

// OHLCVResponse represents the response for OHLCV data requests
type OHLCVResponse struct {
	Symbol    string          `json:"symbol"`
	Timeframe string          `json:"timeframe"`
	Count     int             `json:"count"`
	Data      []*models.OHLCV `json:"data"`
}

// OHLCVHistoryResponse represents the response for historical OHLCV requests
type OHLCVHistoryResponse struct {
	Symbol    string          `json:"symbol"`
	Timeframe string          `json:"timeframe"`
	Start     time.Time       `json:"start"`
	End       time.Time       `json:"end"`
	Count     int             `json:"count"`
	Data      []*models.OHLCV `json:"data"`
}

// SymbolRequest represents a request to add symbols
type SymbolRequest struct {
	Symbols []string `json:"symbols" validate:"required,min=1"`
}

// SymbolResponse represents the response for symbol operations
type SymbolResponse struct {
	Symbols []string `json:"symbols"`
	Count   int      `json:"count"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    string                 `json:"version"`
	Components map[string]interface{} `json:"components"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error         string    `json:"error"`
	Message       string    `json:"message"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}
