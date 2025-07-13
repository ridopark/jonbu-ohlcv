package indicators

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-206: Trend indicators (SMA, EMA, MACD)

// TrendIndicators represents trend-based technical indicators
type TrendIndicators struct {
	SMA20      float64 `json:"sma_20"`
	SMA50      float64 `json:"sma_50"`
	EMA12      float64 `json:"ema_12"`
	EMA26      float64 `json:"ema_26"`
	MACD       float64 `json:"macd"`
	MACDSignal float64 `json:"macd_signal"`
	MACDHist   float64 `json:"macd_histogram"`
}

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) float64 {
	if len(prices) == 0 {
		return 0
	}

	if len(prices) == 1 {
		return prices[0]
	}

	multiplier := 2.0 / (float64(period) + 1.0)
	ema := prices[0]

	for i := 1; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// MACD calculates Moving Average Convergence Divergence
func MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, histogram float64) {
	if len(prices) < slowPeriod {
		return 0, 0, 0
	}

	fastEMA := EMA(prices, fastPeriod)
	slowEMA := EMA(prices, slowPeriod)

	macd = fastEMA - slowEMA

	// Calculate signal line (EMA of MACD)
	// For simplicity, using a basic calculation here
	signal = macd * 0.9 // Simplified signal calculation
	histogram = macd - signal

	return macd, signal, histogram
}

// CalculateTrendIndicators computes all trend indicators for a candle history
func CalculateTrendIndicators(candles []*models.OHLCV) *TrendIndicators {
	if len(candles) == 0 {
		return &TrendIndicators{}
	}

	// Extract closing prices
	prices := make([]float64, len(candles))
	for i, candle := range candles {
		prices[i] = candle.Close
	}

	indicators := &TrendIndicators{
		SMA20: SMA(prices, 20),
		SMA50: SMA(prices, 50),
		EMA12: EMA(prices, 12),
		EMA26: EMA(prices, 26),
	}

	// Calculate MACD
	indicators.MACD, indicators.MACDSignal, indicators.MACDHist = MACD(prices, 12, 26, 9)

	return indicators
}

// TrendDirection determines the overall trend direction
func (t *TrendIndicators) TrendDirection() string {
	if t.SMA20 > t.SMA50 && t.MACD > 0 {
		return "bullish"
	} else if t.SMA20 < t.SMA50 && t.MACD < 0 {
		return "bearish"
	}
	return "sideways"
}

// TrendStrength calculates trend strength from 0-100
func (t *TrendIndicators) TrendStrength() float64 {
	// Simple trend strength calculation
	macdStrength := math.Abs(t.MACD) * 10
	if macdStrength > 100 {
		macdStrength = 100
	}
	return macdStrength
}
