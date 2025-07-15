package models

import (
	"time"
)

// REQ-071: OHLCV MUST include symbol, timestamp, open, high, low, close, volume
// REQ-072: Timestamps MUST be in market timezone (America/New_York)
// REQ-073: Prices MUST be stored as decimal/float64 with proper precision
// REQ-074: Volume MUST be stored as integer/int64
// REQ-075: All fields MUST have appropriate JSON and database tags
type OHLCV struct {
	ID        int64     `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol" validate:"required,uppercase"`
	Timestamp time.Time `json:"timestamp" db:"timestamp" validate:"required"`
	Open      float64   `json:"open" db:"open" validate:"required,gt=0"`
	High      float64   `json:"high" db:"high" validate:"required,gt=0"`
	Low       float64   `json:"low" db:"low" validate:"required,gt=0"`
	Close     float64   `json:"close" db:"close" validate:"required,gt=0"`
	Volume    int64     `json:"volume" db:"volume" validate:"required,gte=0"`
	Timeframe string    `json:"timeframe" db:"timeframe" validate:"required,oneof=1m 5m 15m 1h 4h 1d"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Candle represents a complete OHLCV candle for streaming
type Candle struct {
	Symbol     string    `json:"symbol" validate:"required"`
	Timestamp  time.Time `json:"timestamp" validate:"required"`
	Open       float64   `json:"open" validate:"required,gt=0"`
	High       float64   `json:"high" validate:"required,gt=0"`
	Low        float64   `json:"low" validate:"required,gt=0"`
	Close      float64   `json:"close" validate:"required,gt=0"`
	Volume     int64     `json:"volume" validate:"required,gte=0"`
	Interval   string    `json:"interval" validate:"required"`
	LastUpdate time.Time `json:"last_update"`
}

// MarketEvent represents incoming market data events
type MarketEvent struct {
	Symbol    string    `json:"symbol" validate:"required"`
	Price     float64   `json:"price" validate:"required,gt=0"`
	Volume    int64     `json:"volume" validate:"required,gte=0"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
	Type      string    `json:"type" validate:"required,oneof=trade quote bar"`

	// OHLC fields for bar events (optional, only used when Type == "bar")
	Open  float64 `json:"open,omitempty"`
	High  float64 `json:"high,omitempty"`
	Low   float64 `json:"low,omitempty"`
	Close float64 `json:"close,omitempty"`
}

// REQ-076: Market hours validation
func (o *OHLCV) IsMarketHours() bool {
	// Market hours: 9:30 AM - 4:00 PM EST
	hour := o.Timestamp.Hour()
	minute := o.Timestamp.Minute()

	if hour < 9 || hour > 16 {
		return false
	}
	if hour == 9 && minute < 30 {
		return false
	}
	if hour == 16 && minute > 0 {
		return false
	}

	return true
}

// REQ-077: Pre-market and after-hours detection
func (o *OHLCV) IsExtendedHours() bool {
	hour := o.Timestamp.Hour()
	return (hour >= 4 && hour < 9) || (hour >= 16 && hour <= 20)
}

// Validate performs business logic validation
func (c *Candle) Validate() error {
	// REQ-041: Input validation
	if c.Symbol == "" {
		return ErrInvalidSymbol
	}
	if c.High < c.Low {
		return ErrInvalidPriceRange
	}
	if c.Open < 0 || c.Close < 0 {
		return ErrNegativePrice
	}
	if c.Volume < 0 {
		return ErrNegativeVolume
	}

	return nil
}
