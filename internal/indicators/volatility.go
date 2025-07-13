package indicators

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-208: Volatility indicators (Bollinger Bands, ATR)

// VolatilityIndicators represents volatility-based technical indicators
type VolatilityIndicators struct {
	BollingerUpper  float64 `json:"bollinger_upper"`
	BollingerMiddle float64 `json:"bollinger_middle"`
	BollingerLower  float64 `json:"bollinger_lower"`
	ATR             float64 `json:"atr"`
	StdDev          float64 `json:"standard_deviation"`
	VolatilityRatio float64 `json:"volatility_ratio"`
}

// StandardDeviation calculates standard deviation of prices
func StandardDeviation(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	mean := sum / float64(period)

	// Calculate variance
	variance := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		variance += math.Pow(prices[i]-mean, 2)
	}
	variance /= float64(period)

	return math.Sqrt(variance)
}

// BollingerBands calculates Bollinger Bands
func BollingerBands(prices []float64, period int, stdDevMultiplier float64) (upper, middle, lower float64) {
	if len(prices) < period {
		return 0, 0, 0
	}

	middle = SMA(prices, period)
	stdDev := StandardDeviation(prices, period)

	upper = middle + (stdDev * stdDevMultiplier)
	lower = middle - (stdDev * stdDevMultiplier)

	return upper, middle, lower
}

// TrueRange calculates True Range for a single candle
func TrueRange(current, previous *models.OHLCV) float64 {
	if previous == nil {
		return current.High - current.Low
	}

	tr1 := current.High - current.Low
	tr2 := math.Abs(current.High - previous.Close)
	tr3 := math.Abs(current.Low - previous.Close)

	return math.Max(tr1, math.Max(tr2, tr3))
}

// ATR calculates Average True Range
func ATR(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	trSum := 0.0
	for i := len(candles) - period; i < len(candles); i++ {
		var previous *models.OHLCV
		if i > 0 {
			previous = candles[i-1]
		}
		trSum += TrueRange(candles[i], previous)
	}

	return trSum / float64(period)
}

// CalculateVolatilityIndicators computes all volatility indicators for a candle history
func CalculateVolatilityIndicators(candles []*models.OHLCV) *VolatilityIndicators {
	if len(candles) == 0 {
		return &VolatilityIndicators{}
	}

	// Extract closing prices
	closes := make([]float64, len(candles))
	for i, candle := range candles {
		closes[i] = candle.Close
	}

	indicators := &VolatilityIndicators{
		ATR:    ATR(candles, 14),
		StdDev: StandardDeviation(closes, 20),
	}

	// Calculate Bollinger Bands
	indicators.BollingerUpper, indicators.BollingerMiddle, indicators.BollingerLower =
		BollingerBands(closes, 20, 2.0)

	// Calculate volatility ratio (current vs historical)
	if len(candles) >= 2 {
		currentVolatility := math.Abs(candles[len(candles)-1].Close - candles[len(candles)-2].Close)
		if indicators.ATR > 0 {
			indicators.VolatilityRatio = currentVolatility / indicators.ATR
		}
	}

	return indicators
}

// BollingerPosition returns position relative to Bollinger Bands (0-1)
func (v *VolatilityIndicators) BollingerPosition(currentPrice float64) float64 {
	if v.BollingerUpper == v.BollingerLower {
		return 0.5
	}

	position := (currentPrice - v.BollingerLower) / (v.BollingerUpper - v.BollingerLower)

	// Clamp between 0 and 1
	if position < 0 {
		position = 0
	} else if position > 1 {
		position = 1
	}

	return position
}

// VolatilityLevel returns volatility level assessment
func (v *VolatilityIndicators) VolatilityLevel() string {
	if v.VolatilityRatio > 1.5 {
		return "high"
	} else if v.VolatilityRatio < 0.5 {
		return "low"
	}
	return "normal"
}

// IsNearBollingerBands checks if price is near upper or lower bands
func (v *VolatilityIndicators) IsNearBollingerBands(currentPrice float64) (nearUpper, nearLower bool) {
	position := v.BollingerPosition(currentPrice)

	nearUpper = position > 0.9
	nearLower = position < 0.1

	return nearUpper, nearLower
}
