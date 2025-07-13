package indicators

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-207: Momentum indicators (RSI, Stochastic, Williams %R)

// MomentumIndicators represents momentum-based technical indicators
type MomentumIndicators struct {
	RSI         float64 `json:"rsi"`
	StochasticK float64 `json:"stochastic_k"`
	StochasticD float64 `json:"stochastic_d"`
	WilliamsR   float64 `json:"williams_r"`
	ROC         float64 `json:"roc"` // Rate of Change
}

// RSI calculates Relative Strength Index
func RSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50 // Neutral RSI
	}

	gains := 0.0
	losses := 0.0

	// Calculate initial gains and losses
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// Stochastic calculates Stochastic Oscillator %K and %D
func Stochastic(highs, lows, closes []float64, kPeriod, dPeriod int) (k, d float64) {
	if len(closes) < kPeriod {
		return 50, 50 // Neutral values
	}

	// Find highest high and lowest low over the period
	highestHigh := highs[len(highs)-kPeriod]
	lowestLow := lows[len(lows)-kPeriod]

	for i := len(highs) - kPeriod; i < len(highs); i++ {
		if highs[i] > highestHigh {
			highestHigh = highs[i]
		}
		if lows[i] < lowestLow {
			lowestLow = lows[i]
		}
	}

	currentClose := closes[len(closes)-1]

	if highestHigh == lowestLow {
		k = 50
	} else {
		k = ((currentClose - lowestLow) / (highestHigh - lowestLow)) * 100
	}

	// %D is typically a 3-period SMA of %K (simplified here)
	d = k * 0.9 // Simplified calculation

	return k, d
}

// WilliamsR calculates Williams %R
func WilliamsR(highs, lows, closes []float64, period int) float64 {
	if len(closes) < period {
		return -50 // Neutral value
	}

	// Find highest high and lowest low over the period
	highestHigh := highs[len(highs)-period]
	lowestLow := lows[len(lows)-period]

	for i := len(highs) - period; i < len(highs); i++ {
		if highs[i] > highestHigh {
			highestHigh = highs[i]
		}
		if lows[i] < lowestLow {
			lowestLow = lows[i]
		}
	}

	currentClose := closes[len(closes)-1]

	if highestHigh == lowestLow {
		return -50
	}

	return ((highestHigh - currentClose) / (highestHigh - lowestLow)) * -100
}

// ROC calculates Rate of Change
func ROC(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	currentPrice := prices[len(prices)-1]
	pastPrice := prices[len(prices)-1-period]

	if pastPrice == 0 {
		return 0
	}

	return ((currentPrice - pastPrice) / pastPrice) * 100
}

// CalculateMomentumIndicators computes all momentum indicators for a candle history
func CalculateMomentumIndicators(candles []*models.OHLCV) *MomentumIndicators {
	if len(candles) == 0 {
		return &MomentumIndicators{}
	}

	// Extract price arrays
	closes := make([]float64, len(candles))
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))

	for i, candle := range candles {
		closes[i] = candle.Close
		highs[i] = candle.High
		lows[i] = candle.Low
	}

	indicators := &MomentumIndicators{
		RSI:       RSI(closes, 14),
		WilliamsR: WilliamsR(highs, lows, closes, 14),
		ROC:       ROC(closes, 10),
	}

	// Calculate Stochastic
	indicators.StochasticK, indicators.StochasticD = Stochastic(highs, lows, closes, 14, 3)

	return indicators
}

// IsOverbought checks if momentum indicators suggest overbought conditions
func (m *MomentumIndicators) IsOverbought() bool {
	return m.RSI > 70 || m.StochasticK > 80 || m.WilliamsR > -20
}

// IsOversold checks if momentum indicators suggest oversold conditions
func (m *MomentumIndicators) IsOversold() bool {
	return m.RSI < 30 || m.StochasticK < 20 || m.WilliamsR < -80
}

// MomentumSignal returns overall momentum signal
func (m *MomentumIndicators) MomentumSignal() string {
	if m.IsOverbought() {
		return "overbought"
	} else if m.IsOversold() {
		return "oversold"
	}
	return "neutral"
}
