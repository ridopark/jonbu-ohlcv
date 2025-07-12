package models

import "errors"

// REQ-093: Meaningful error messages
var (
	ErrInvalidSymbol     = errors.New("invalid symbol: must be non-empty uppercase letters only")
	ErrInvalidPriceRange = errors.New("invalid price range: high must be greater than or equal to low")
	ErrNegativePrice     = errors.New("invalid price: prices cannot be negative")
	ErrNegativeVolume    = errors.New("invalid volume: volume cannot be negative")
	ErrInvalidTimeframe  = errors.New("invalid timeframe: must be one of 1m, 5m, 15m, 1h, 4h, 1d")
	ErrInvalidTimestamp  = errors.New("invalid timestamp: timestamp cannot be zero")
)

// MarketDataError represents validation errors for market data
type MarketDataError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *MarketDataError) Error() string {
	return e.Message
}
