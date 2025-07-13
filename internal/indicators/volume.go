package indicators

import (
	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-209: Volume indicators (Volume MA, VWAP, OBV)

// VolumeIndicators represents volume-based technical indicators
type VolumeIndicators struct {
	VolumeMA    float64 `json:"volume_ma"`
	VWAP        float64 `json:"vwap"`
	OBV         float64 `json:"obv"`
	VolumeRatio float64 `json:"volume_ratio"`
	AccDist     float64 `json:"accumulation_distribution"`
}

// VolumeMA calculates Volume Moving Average
func VolumeMA(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}

	sum := int64(0)
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Volume
	}

	return float64(sum) / float64(period)
}

// VWAP calculates Volume Weighted Average Price
func VWAP(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	var totalVolume int64
	var totalPriceVolume float64

	for _, candle := range candles {
		typicalPrice := (candle.High + candle.Low + candle.Close) / 3.0
		totalPriceVolume += typicalPrice * float64(candle.Volume)
		totalVolume += candle.Volume
	}

	if totalVolume == 0 {
		return 0
	}

	return totalPriceVolume / float64(totalVolume)
}

// OBV calculates On-Balance Volume
func OBV(candles []*models.OHLCV) float64 {
	if len(candles) < 2 {
		return 0
	}

	obv := float64(candles[0].Volume)

	for i := 1; i < len(candles); i++ {
		if candles[i].Close > candles[i-1].Close {
			obv += float64(candles[i].Volume)
		} else if candles[i].Close < candles[i-1].Close {
			obv -= float64(candles[i].Volume)
		}
		// If close prices are equal, OBV remains unchanged
	}

	return obv
}

// AccumulationDistribution calculates Accumulation/Distribution Line
func AccumulationDistribution(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	var adLine float64

	for _, candle := range candles {
		if candle.High == candle.Low {
			continue // Avoid division by zero
		}

		// Money Flow Multiplier
		mfm := ((candle.Close - candle.Low) - (candle.High - candle.Close)) / (candle.High - candle.Low)

		// Money Flow Volume
		mfv := mfm * float64(candle.Volume)

		adLine += mfv
	}

	return adLine
}

// CalculateVolumeIndicators computes all volume indicators for a candle history
func CalculateVolumeIndicators(candles []*models.OHLCV) *VolumeIndicators {
	if len(candles) == 0 {
		return &VolumeIndicators{}
	}

	indicators := &VolumeIndicators{
		VolumeMA: VolumeMA(candles, 20),
		VWAP:     VWAP(candles),
		OBV:      OBV(candles),
		AccDist:  AccumulationDistribution(candles),
	}

	// Calculate volume ratio (current vs average)
	if len(candles) > 0 && indicators.VolumeMA > 0 {
		currentVolume := float64(candles[len(candles)-1].Volume)
		indicators.VolumeRatio = currentVolume / indicators.VolumeMA
	}

	return indicators
}

// VolumeSignal returns volume-based signal
func (v *VolumeIndicators) VolumeSignal() string {
	if v.VolumeRatio > 1.5 {
		return "high_volume"
	} else if v.VolumeRatio < 0.5 {
		return "low_volume"
	}
	return "normal_volume"
}

// IsAboveVWAP checks if current price is above VWAP
func (v *VolumeIndicators) IsAboveVWAP(currentPrice float64) bool {
	return v.VWAP > 0 && currentPrice > v.VWAP
}

// VolumeConfirmation checks if volume confirms price movement
func (v *VolumeIndicators) VolumeConfirmation() bool {
	// High volume with positive OBV change suggests strong movement
	return v.VolumeRatio > 1.2 && v.OBV > 0
}

// AccumulationSignal returns accumulation/distribution signal
func (v *VolumeIndicators) AccumulationSignal() string {
	if v.AccDist > 0 {
		return "accumulation"
	} else if v.AccDist < 0 {
		return "distribution"
	}
	return "neutral"
}
