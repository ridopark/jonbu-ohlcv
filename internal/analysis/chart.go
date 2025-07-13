package analysis

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-212: Chart patterns (breakouts, reversals, continuations)

// ChartPatternAnalyzer detects chart patterns
type ChartPatternAnalyzer struct {
	minPatternLength int     // Minimum candles for pattern
	tolerance        float64 // Price tolerance for pattern recognition
}

// NewChartPatternAnalyzer creates a new chart pattern analyzer
func NewChartPatternAnalyzer() *ChartPatternAnalyzer {
	return &ChartPatternAnalyzer{
		minPatternLength: 10,   // Minimum 10 candles
		tolerance:        0.02, // 2% tolerance
	}
}

// DetectPatterns detects chart patterns
func (cpa *ChartPatternAnalyzer) DetectPatterns(candles []*models.OHLCV) []ChartPatternResult {
	if len(candles) < cpa.minPatternLength {
		return nil
	}

	patterns := make([]ChartPatternResult, 0)

	// Detect various chart patterns
	if breakout := cpa.detectBreakout(candles); breakout != nil {
		patterns = append(patterns, *breakout)
	}

	if triangle := cpa.detectTriangle(candles); triangle != nil {
		patterns = append(patterns, *triangle)
	}

	if headShoulders := cpa.detectHeadAndShoulders(candles); headShoulders != nil {
		patterns = append(patterns, *headShoulders)
	}

	return patterns
}

// ChartPatternResult represents a detected chart pattern
type ChartPatternResult struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Signal     string  `json:"signal"`
	Target     float64 `json:"target,omitempty"`
	StopLoss   float64 `json:"stop_loss,omitempty"`
	Timeframe  string  `json:"timeframe"`
	Status     string  `json:"status"`
}

func (cpa *ChartPatternAnalyzer) detectBreakout(candles []*models.OHLCV) *ChartPatternResult {
	if len(candles) < 20 {
		return nil
	}

	// Look for consolidation followed by breakout
	recent := candles[len(candles)-10:]
	consolidation := candles[len(candles)-20 : len(candles)-10]

	// Calculate consolidation range
	high := consolidation[0].High
	low := consolidation[0].Low

	for _, candle := range consolidation {
		if candle.High > high {
			high = candle.High
		}
		if candle.Low < low {
			low = candle.Low
		}
	}

	rangeSize := high - low

	// Check if recent candles break out
	for _, candle := range recent {
		if candle.Close > high {
			return &ChartPatternResult{
				Type:       "breakout",
				Name:       "Bullish Breakout",
				Confidence: 75.0,
				Signal:     "bullish",
				Target:     candle.Close + rangeSize,
				StopLoss:   high * 0.98,
				Timeframe:  "medium",
				Status:     "confirmed",
			}
		}
		if candle.Close < low {
			return &ChartPatternResult{
				Type:       "breakout",
				Name:       "Bearish Breakdown",
				Confidence: 75.0,
				Signal:     "bearish",
				Target:     candle.Close - rangeSize,
				StopLoss:   low * 1.02,
				Timeframe:  "medium",
				Status:     "confirmed",
			}
		}
	}

	return nil
}

func (cpa *ChartPatternAnalyzer) detectTriangle(candles []*models.OHLCV) *ChartPatternResult {
	if len(candles) < 30 {
		return nil
	}

	// Find recent highs and lows
	highs := make([]float64, 0)
	lows := make([]float64, 0)

	for i := 5; i < len(candles)-5; i++ {
		isHigh := true
		isLow := true

		// Check if it's a local high
		for j := i - 5; j <= i+5; j++ {
			if j != i && candles[j].High >= candles[i].High {
				isHigh = false
			}
		}

		// Check if it's a local low
		for j := i - 5; j <= i+5; j++ {
			if j != i && candles[j].Low <= candles[i].Low {
				isLow = false
			}
		}

		if isHigh {
			highs = append(highs, candles[i].High)
		}
		if isLow {
			lows = append(lows, candles[i].Low)
		}
	}

	// Check for converging trend lines (triangle pattern)
	if len(highs) >= 2 && len(lows) >= 2 {
		// Simplified triangle detection
		highSlope := (highs[len(highs)-1] - highs[0]) / float64(len(highs))
		lowSlope := (lows[len(lows)-1] - lows[0]) / float64(len(lows))

		// Converging lines indicate triangle
		if math.Abs(highSlope-lowSlope) > 0.1 && math.Abs(highSlope) < 1.0 && math.Abs(lowSlope) < 1.0 {
			currentPrice := candles[len(candles)-1].Close

			return &ChartPatternResult{
				Type:       "triangle",
				Name:       "Symmetrical Triangle",
				Confidence: 65.0,
				Signal:     "neutral",
				Target:     currentPrice * 1.05, // 5% move expected
				StopLoss:   currentPrice * 0.95,
				Timeframe:  "medium",
				Status:     "forming",
			}
		}
	}

	return nil
}

func (cpa *ChartPatternAnalyzer) detectHeadAndShoulders(candles []*models.OHLCV) *ChartPatternResult {
	if len(candles) < 50 {
		return nil
	}

	// Find three prominent peaks
	peaks := make([]struct {
		index int
		price float64
	}, 0)

	for i := 10; i < len(candles)-10; i++ {
		isPeak := true

		// Check if it's a significant peak
		for j := i - 10; j <= i+10; j++ {
			if j != i && candles[j].High >= candles[i].High {
				isPeak = false
				break
			}
		}

		if isPeak {
			peaks = append(peaks, struct {
				index int
				price float64
			}{i, candles[i].High})
		}
	}

	// Need at least 3 peaks for head and shoulders
	if len(peaks) >= 3 {
		// Find the highest peak (head) and two shoulders
		for i := 1; i < len(peaks)-1; i++ {
			leftShoulder := peaks[i-1]
			head := peaks[i]
			rightShoulder := peaks[i+1]

			// Head should be higher than both shoulders
			if head.price > leftShoulder.price && head.price > rightShoulder.price {
				// Shoulders should be roughly equal
				shoulderRatio := leftShoulder.price / rightShoulder.price

				if shoulderRatio > 0.95 && shoulderRatio < 1.05 {
					neckline := math.Min(leftShoulder.price, rightShoulder.price) * 0.98

					return &ChartPatternResult{
						Type:       "head_shoulders",
						Name:       "Head and Shoulders",
						Confidence: 80.0,
						Signal:     "bearish",
						Target:     neckline - (head.price - neckline),
						StopLoss:   head.price,
						Timeframe:  "long",
						Status:     "confirmed",
					}
				}
			}
		}
	}

	return nil
}

// ChartPattern represents a detected chart pattern
type ChartPattern struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`      // breakout, reversal, continuation
	Direction   string  `json:"direction"` // bullish, bearish, neutral
	Strength    float64 `json:"strength"`  // 0-100
	TargetPrice float64 `json:"target_price,omitempty"`
	Description string  `json:"description"`
}

// ChartPatternDetector detects chart patterns
type ChartPatternDetector struct{}

// NewChartPatternDetector creates a new chart pattern detector
func NewChartPatternDetector() *ChartPatternDetector {
	return &ChartPatternDetector{}
}

// DetectPatterns identifies chart patterns in candle data
func (cpd *ChartPatternDetector) DetectPatterns(candles []*models.OHLCV) []*ChartPattern {
	if len(candles) < 10 {
		return nil
	}

	var patterns []*ChartPattern

	// Support and resistance levels
	support, resistance := cpd.findSupportResistance(candles)

	// Breakout patterns
	if pattern := cpd.detectBreakout(candles, support, resistance); pattern != nil {
		patterns = append(patterns, pattern)
	}

	// Triangle patterns
	if pattern := cpd.detectTriangle(candles); pattern != nil {
		patterns = append(patterns, pattern)
	}

	// Head and shoulders
	if pattern := cpd.detectHeadAndShoulders(candles); pattern != nil {
		patterns = append(patterns, pattern)
	}

	// Double top/bottom
	if pattern := cpd.detectDoubleTopBottom(candles); pattern != nil {
		patterns = append(patterns, pattern)
	}

	return patterns
}

// findSupportResistance identifies key support and resistance levels
func (cpd *ChartPatternDetector) findSupportResistance(candles []*models.OHLCV) (support, resistance float64) {
	if len(candles) < 5 {
		return 0, 0
	}

	// Find recent lows and highs
	lows := make([]float64, 0)
	highs := make([]float64, 0)

	// Look for swing lows and highs
	for i := 2; i < len(candles)-2; i++ {
		// Swing low: lower than 2 candles on each side
		if candles[i].Low < candles[i-1].Low && candles[i].Low < candles[i-2].Low &&
			candles[i].Low < candles[i+1].Low && candles[i].Low < candles[i+2].Low {
			lows = append(lows, candles[i].Low)
		}

		// Swing high: higher than 2 candles on each side
		if candles[i].High > candles[i-1].High && candles[i].High > candles[i-2].High &&
			candles[i].High > candles[i+1].High && candles[i].High > candles[i+2].High {
			highs = append(highs, candles[i].High)
		}
	}

	// Find most significant levels (clustering)
	if len(lows) > 0 {
		support = cpd.findMostSignificantLevel(lows)
	}
	if len(highs) > 0 {
		resistance = cpd.findMostSignificantLevel(highs)
	}

	return support, resistance
}

// findMostSignificantLevel finds the most significant level from a slice of levels
func (cpd *ChartPatternDetector) findMostSignificantLevel(levels []float64) float64 {
	if len(levels) == 0 {
		return 0
	}

	// Simple approach: return the level that appears most frequently within 1% range
	tolerance := 0.01 // 1%
	maxCount := 0
	significantLevel := levels[0]

	for _, level := range levels {
		count := 0
		for _, otherLevel := range levels {
			if math.Abs(level-otherLevel)/level < tolerance {
				count++
			}
		}
		if count > maxCount {
			maxCount = count
			significantLevel = level
		}
	}

	return significantLevel
}

// detectBreakout detects breakout patterns
func (cpd *ChartPatternDetector) detectBreakout(candles []*models.OHLCV, support, resistance float64) *ChartPattern {
	if len(candles) < 2 || support == 0 || resistance == 0 {
		return nil
	}

	current := candles[len(candles)-1]
	previous := candles[len(candles)-2]

	// Volume threshold for confirming breakout
	recentVolume := float64(current.Volume)
	avgVolume := cpd.calculateAverageVolume(candles, 20)

	// Resistance breakout
	if previous.Close <= resistance && current.Close > resistance && recentVolume > avgVolume*1.5 {
		target := resistance + (resistance - support)
		return &ChartPattern{
			Name:        "Resistance Breakout",
			Type:        "breakout",
			Direction:   "bullish",
			Strength:    75.0,
			TargetPrice: target,
			Description: "Price broke above resistance with high volume",
		}
	}

	// Support breakdown
	if previous.Close >= support && current.Close < support && recentVolume > avgVolume*1.5 {
		target := support - (resistance - support)
		return &ChartPattern{
			Name:        "Support Breakdown",
			Type:        "breakout",
			Direction:   "bearish",
			Strength:    75.0,
			TargetPrice: target,
			Description: "Price broke below support with high volume",
		}
	}

	return nil
}

// detectTriangle detects triangle patterns
func (cpd *ChartPatternDetector) detectTriangle(candles []*models.OHLCV) *ChartPattern {
	if len(candles) < 20 {
		return nil
	}

	// Look for converging trend lines
	recent := candles[len(candles)-20:]

	// Find highs and lows for trend lines
	highs := make([]float64, 0)
	lows := make([]float64, 0)

	for i := 1; i < len(recent)-1; i++ {
		if recent[i].High > recent[i-1].High && recent[i].High > recent[i+1].High {
			highs = append(highs, recent[i].High)
		}
		if recent[i].Low < recent[i-1].Low && recent[i].Low < recent[i+1].Low {
			lows = append(lows, recent[i].Low)
		}
	}

	if len(highs) < 2 || len(lows) < 2 {
		return nil
	}

	// Calculate trend line slopes
	highSlope := cpd.calculateSlope(highs)
	lowSlope := cpd.calculateSlope(lows)

	// Ascending triangle: flat resistance, rising support
	if math.Abs(highSlope) < 0.001 && lowSlope > 0.001 {
		return &ChartPattern{
			Name:        "Ascending Triangle",
			Type:        "continuation",
			Direction:   "bullish",
			Strength:    65.0,
			Description: "Bullish continuation pattern with rising support",
		}
	}

	// Descending triangle: falling resistance, flat support
	if highSlope < -0.001 && math.Abs(lowSlope) < 0.001 {
		return &ChartPattern{
			Name:        "Descending Triangle",
			Type:        "continuation",
			Direction:   "bearish",
			Strength:    65.0,
			Description: "Bearish continuation pattern with falling resistance",
		}
	}

	// Symmetrical triangle: converging lines
	if highSlope < -0.001 && lowSlope > 0.001 && math.Abs(highSlope+lowSlope) < 0.002 {
		return &ChartPattern{
			Name:        "Symmetrical Triangle",
			Type:        "continuation",
			Direction:   "neutral",
			Strength:    50.0,
			Description: "Neutral continuation pattern, awaiting breakout direction",
		}
	}

	return nil
}

// detectHeadAndShoulders detects head and shoulders patterns
func (cpd *ChartPatternDetector) detectHeadAndShoulders(candles []*models.OHLCV) *ChartPattern {
	if len(candles) < 30 {
		return nil
	}

	// Look for three peaks pattern
	recent := candles[len(candles)-30:]
	peaks := make([]struct {
		index int
		value float64
	}, 0)

	// Find peaks
	for i := 2; i < len(recent)-2; i++ {
		if recent[i].High > recent[i-1].High && recent[i].High > recent[i-2].High &&
			recent[i].High > recent[i+1].High && recent[i].High > recent[i+2].High {
			peaks = append(peaks, struct {
				index int
				value float64
			}{i, recent[i].High})
		}
	}

	if len(peaks) < 3 {
		return nil
	}

	// Check for head and shoulders pattern (middle peak higher than others)
	for i := 1; i < len(peaks)-1; i++ {
		leftShoulder := peaks[i-1]
		head := peaks[i]
		rightShoulder := peaks[i+1]

		// Head should be higher than both shoulders
		if head.value > leftShoulder.value && head.value > rightShoulder.value {
			// Shoulders should be roughly equal (within 5%)
			shoulderDiff := math.Abs(leftShoulder.value-rightShoulder.value) / leftShoulder.value
			if shoulderDiff < 0.05 {
				return &ChartPattern{
					Name:        "Head and Shoulders",
					Type:        "reversal",
					Direction:   "bearish",
					Strength:    80.0,
					Description: "Strong bearish reversal pattern",
				}
			}
		}
	}

	return nil
}

// detectDoubleTopBottom detects double top/bottom patterns
func (cpd *ChartPatternDetector) detectDoubleTopBottom(candles []*models.OHLCV) *ChartPattern {
	if len(candles) < 20 {
		return nil
	}

	recent := candles[len(candles)-20:]

	// Find significant peaks and troughs
	peaks := make([]float64, 0)
	troughs := make([]float64, 0)

	for i := 2; i < len(recent)-2; i++ {
		// Peak
		if recent[i].High > recent[i-1].High && recent[i].High > recent[i-2].High &&
			recent[i].High > recent[i+1].High && recent[i].High > recent[i+2].High {
			peaks = append(peaks, recent[i].High)
		}

		// Trough
		if recent[i].Low < recent[i-1].Low && recent[i].Low < recent[i-2].Low &&
			recent[i].Low < recent[i+1].Low && recent[i].Low < recent[i+2].Low {
			troughs = append(troughs, recent[i].Low)
		}
	}

	// Double top: two peaks at similar levels
	if len(peaks) >= 2 {
		for i := 0; i < len(peaks)-1; i++ {
			for j := i + 1; j < len(peaks); j++ {
				diff := math.Abs(peaks[i]-peaks[j]) / peaks[i]
				if diff < 0.03 { // Within 3%
					return &ChartPattern{
						Name:        "Double Top",
						Type:        "reversal",
						Direction:   "bearish",
						Strength:    70.0,
						Description: "Bearish reversal pattern with two peaks",
					}
				}
			}
		}
	}

	// Double bottom: two troughs at similar levels
	if len(troughs) >= 2 {
		for i := 0; i < len(troughs)-1; i++ {
			for j := i + 1; j < len(troughs); j++ {
				diff := math.Abs(troughs[i]-troughs[j]) / troughs[i]
				if diff < 0.03 { // Within 3%
					return &ChartPattern{
						Name:        "Double Bottom",
						Type:        "reversal",
						Direction:   "bullish",
						Strength:    70.0,
						Description: "Bullish reversal pattern with two troughs",
					}
				}
			}
		}
	}

	return nil
}

// calculateAverageVolume calculates average volume over a period
func (cpd *ChartPatternDetector) calculateAverageVolume(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		period = len(candles)
	}

	sum := int64(0)
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Volume
	}

	return float64(sum) / float64(period)
}

// calculateSlope calculates the slope of a series of values
func (cpd *ChartPatternDetector) calculateSlope(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX := (n - 1) * n / 2 // Sum of indices 0, 1, 2, ..., n-1
	sumY := 0.0
	sumXY := 0.0
	sumX2 := (n - 1) * n * (2*n - 1) / 6 // Sum of squares of indices

	for i, value := range values {
		sumY += value
		sumXY += float64(i) * value
	}

	// Linear regression slope: (n*sumXY - sumX*sumY) / (n*sumX2 - sumX^2)
	numerator := n*sumXY - sumX*sumY
	denominator := n*sumX2 - sumX*sumX

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
