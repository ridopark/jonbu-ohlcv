package analysis

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-211: Candlestick patterns (doji, hammer, shooting star, etc.)

// CandlestickAnalyzer detects candlestick patterns
type CandlestickAnalyzer struct {
	// Configuration for pattern detection
	minBodyPercent float64 // Minimum body size as percentage of range
	wickRatio      float64 // Wick to body ratio for specific patterns
}

// NewCandlestickAnalyzer creates a new candlestick analyzer
func NewCandlestickAnalyzer() *CandlestickAnalyzer {
	return &CandlestickAnalyzer{
		minBodyPercent: 0.1, // 10% minimum body size
		wickRatio:      2.0, // 2:1 wick to body ratio
	}
}

// DetectPatterns detects candlestick patterns in a candle sequence
func (ca *CandlestickAnalyzer) DetectPatterns(candles []*models.OHLCV) []string {
	if len(candles) < 3 {
		return nil
	}

	patterns := make([]string, 0)

	// Single candle patterns
	current := candles[len(candles)-1]

	if ca.isDoji(current) {
		patterns = append(patterns, "doji")
	}

	if ca.isHammer(current) {
		patterns = append(patterns, "hammer")
	}

	if ca.isShootingStar(current) {
		patterns = append(patterns, "shooting_star")
	}

	// Multi-candle patterns (need at least 2 candles)
	if len(candles) >= 2 {
		prev := candles[len(candles)-2]

		if ca.isBullishEngulfing(prev, current) {
			patterns = append(patterns, "engulfing_bullish")
		}

		if ca.isBearishEngulfing(prev, current) {
			patterns = append(patterns, "engulfing_bearish")
		}
	}

	// Three-candle patterns
	if len(candles) >= 3 {
		first := candles[len(candles)-3]
		middle := candles[len(candles)-2]
		last := candles[len(candles)-1]

		if ca.isMorningStar(first, middle, last) {
			patterns = append(patterns, "morning_star")
		}

		if ca.isEveningStar(first, middle, last) {
			patterns = append(patterns, "evening_star")
		}
	}

	return patterns
}

// Single candle pattern detection methods

func (ca *CandlestickAnalyzer) isDoji(candle *models.OHLCV) bool {
	bodySize := math.Abs(candle.Close - candle.Open)
	totalRange := candle.High - candle.Low

	if totalRange == 0 {
		return false
	}

	bodyPercent := bodySize / totalRange
	return bodyPercent < ca.minBodyPercent
}

func (ca *CandlestickAnalyzer) isHammer(candle *models.OHLCV) bool {
	bodySize := math.Abs(candle.Close - candle.Open)
	totalRange := candle.High - candle.Low

	if totalRange == 0 || bodySize == 0 {
		return false
	}

	// Hammer has small body in upper part and long lower shadow
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)

	return lowerShadow > bodySize*ca.wickRatio && upperShadow < bodySize*0.5
}

func (ca *CandlestickAnalyzer) isShootingStar(candle *models.OHLCV) bool {
	bodySize := math.Abs(candle.Close - candle.Open)
	totalRange := candle.High - candle.Low

	if totalRange == 0 || bodySize == 0 {
		return false
	}

	// Shooting star has small body in lower part and long upper shadow
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)

	return upperShadow > bodySize*ca.wickRatio && lowerShadow < bodySize*0.5
}

// Multi-candle pattern detection methods

func (ca *CandlestickAnalyzer) isBullishEngulfing(prev, current *models.OHLCV) bool {
	// Previous candle is bearish, current is bullish and engulfs previous
	if prev.Close >= prev.Open || current.Close <= current.Open {
		return false
	}

	return current.Open < prev.Close && current.Close > prev.Open
}

func (ca *CandlestickAnalyzer) isBearishEngulfing(prev, current *models.OHLCV) bool {
	// Previous candle is bearish, current is bearish and engulfs previous
	if prev.Close <= prev.Open || current.Close >= current.Open {
		return false
	}

	return current.Open > prev.Close && current.Close < prev.Open
}

func (ca *CandlestickAnalyzer) isMorningStar(first, middle, last *models.OHLCV) bool {
	// First candle bearish, middle is doji/small, last is bullish
	if first.Close >= first.Open || last.Close <= last.Open {
		return false
	}

	// Middle candle should be small
	middleBody := math.Abs(middle.Close - middle.Open)
	firstBody := math.Abs(first.Close - first.Open)
	lastBody := math.Abs(last.Close - last.Open)

	if middleBody > firstBody*0.3 || middleBody > lastBody*0.3 {
		return false
	}

	// Last candle should close above middle of first candle
	firstMiddle := (first.Open + first.Close) / 2
	return last.Close > firstMiddle
}

func (ca *CandlestickAnalyzer) isEveningStar(first, middle, last *models.OHLCV) bool {
	// First candle bullish, middle is doji/small, last is bearish
	if first.Close <= first.Open || last.Close >= last.Open {
		return false
	}

	// Middle candle should be small
	middleBody := math.Abs(middle.Close - middle.Open)
	firstBody := math.Abs(first.Close - first.Open)
	lastBody := math.Abs(last.Close - last.Open)

	if middleBody > firstBody*0.3 || middleBody > lastBody*0.3 {
		return false
	}

	// Last candle should close below middle of first candle
	firstMiddle := (first.Open + first.Close) / 2
	return last.Close < firstMiddle
}

// CandlestickPattern represents a detected candlestick pattern
type CandlestickPattern struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`     // bullish, bearish, neutral
	Strength    float64 `json:"strength"` // 0-100
	Description string  `json:"description"`
}

// PatternDetector detects candlestick patterns (legacy interface)
type PatternDetector struct{}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{}
}

// DetectPatterns identifies candlestick patterns in recent candles
func (pd *PatternDetector) DetectPatterns(candles []*models.OHLCV) []*CandlestickPattern {
	if len(candles) == 0 {
		return nil
	}

	var patterns []*CandlestickPattern

	// Single candle patterns
	if len(candles) >= 1 {
		current := candles[len(candles)-1]

		if pattern := pd.detectDoji(current); pattern != nil {
			patterns = append(patterns, pattern)
		}

		if pattern := pd.detectHammer(current); pattern != nil {
			patterns = append(patterns, pattern)
		}

		if pattern := pd.detectShootingStar(current); pattern != nil {
			patterns = append(patterns, pattern)
		}
	}

	// Two candle patterns
	if len(candles) >= 2 {
		current := candles[len(candles)-1]
		previous := candles[len(candles)-2]

		if pattern := pd.detectEngulfing(previous, current); pattern != nil {
			patterns = append(patterns, pattern)
		}
	}

	// Three candle patterns
	if len(candles) >= 3 {
		third := candles[len(candles)-3]
		second := candles[len(candles)-2]
		first := candles[len(candles)-1]

		if pattern := pd.detectMorningStar(third, second, first); pattern != nil {
			patterns = append(patterns, pattern)
		}

		if pattern := pd.detectEveningStar(third, second, first); pattern != nil {
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// detectDoji detects doji patterns
func (pd *PatternDetector) detectDoji(candle *models.OHLCV) *CandlestickPattern {
	bodySize := math.Abs(candle.Close - candle.Open)
	totalSize := candle.High - candle.Low

	if totalSize == 0 {
		return nil
	}

	bodyRatio := bodySize / totalSize

	// Doji: very small body relative to total range
	if bodyRatio < 0.1 {
		return &CandlestickPattern{
			Name:        "Doji",
			Type:        "neutral",
			Strength:    (1 - bodyRatio) * 100,
			Description: "Indecision pattern with small body, potential reversal signal",
		}
	}

	return nil
}

// detectHammer detects hammer patterns
func (pd *PatternDetector) detectHammer(candle *models.OHLCV) *CandlestickPattern {
	bodySize := math.Abs(candle.Close - candle.Open)
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)
	totalSize := candle.High - candle.Low

	if totalSize == 0 {
		return nil
	}

	// Hammer: small body, long lower shadow, small upper shadow
	if bodySize/totalSize < 0.3 && lowerShadow > 2*bodySize && upperShadow < bodySize {
		patternType := "bullish"
		if candle.Close < candle.Open {
			patternType = "bearish"
		}

		strength := (lowerShadow / totalSize) * 100

		return &CandlestickPattern{
			Name:        "Hammer",
			Type:        patternType,
			Strength:    strength,
			Description: "Potential bullish reversal with long lower shadow",
		}
	}

	return nil
}

// detectShootingStar detects shooting star patterns
func (pd *PatternDetector) detectShootingStar(candle *models.OHLCV) *CandlestickPattern {
	bodySize := math.Abs(candle.Close - candle.Open)
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)
	totalSize := candle.High - candle.Low

	if totalSize == 0 {
		return nil
	}

	// Shooting star: small body, long upper shadow, small lower shadow
	if bodySize/totalSize < 0.3 && upperShadow > 2*bodySize && lowerShadow < bodySize {
		strength := (upperShadow / totalSize) * 100

		return &CandlestickPattern{
			Name:        "Shooting Star",
			Type:        "bearish",
			Strength:    strength,
			Description: "Potential bearish reversal with long upper shadow",
		}
	}

	return nil
}

// detectEngulfing detects bullish/bearish engulfing patterns
func (pd *PatternDetector) detectEngulfing(prev, curr *models.OHLCV) *CandlestickPattern {
	prevBody := math.Abs(prev.Close - prev.Open)
	currBody := math.Abs(curr.Close - curr.Open)

	// Current candle must have larger body than previous
	if currBody <= prevBody {
		return nil
	}

	// Bullish engulfing: prev bearish, curr bullish, curr engulfs prev
	if prev.Close < prev.Open && curr.Close > curr.Open &&
		curr.Open < prev.Close && curr.Close > prev.Open {

		strength := (currBody / prevBody) * 50
		if strength > 100 {
			strength = 100
		}

		return &CandlestickPattern{
			Name:        "Bullish Engulfing",
			Type:        "bullish",
			Strength:    strength,
			Description: "Strong bullish reversal pattern",
		}
	}

	// Bearish engulfing: prev bullish, curr bearish, curr engulfs prev
	if prev.Close > prev.Open && curr.Close < curr.Open &&
		curr.Open > prev.Close && curr.Close < prev.Open {

		strength := (currBody / prevBody) * 50
		if strength > 100 {
			strength = 100
		}

		return &CandlestickPattern{
			Name:        "Bearish Engulfing",
			Type:        "bearish",
			Strength:    strength,
			Description: "Strong bearish reversal pattern",
		}
	}

	return nil
}

// detectMorningStar detects morning star patterns
func (pd *PatternDetector) detectMorningStar(first, second, third *models.OHLCV) *CandlestickPattern {
	// First candle: bearish
	// Second candle: small body (doji/spinning top)
	// Third candle: bullish, closes above midpoint of first candle

	if first.Close >= first.Open || third.Close <= third.Open {
		return nil
	}

	firstBody := first.Open - first.Close
	secondBody := math.Abs(second.Close - second.Open)
	thirdBody := third.Close - third.Open

	firstMidpoint := (first.Open + first.Close) / 2

	// Check pattern conditions
	if secondBody < firstBody*0.5 && // Small middle candle
		third.Close > firstMidpoint && // Third closes above first's midpoint
		thirdBody > firstBody*0.3 { // Third has decent size

		return &CandlestickPattern{
			Name:        "Morning Star",
			Type:        "bullish",
			Strength:    75.0,
			Description: "Strong three-candle bullish reversal pattern",
		}
	}

	return nil
}

// detectEveningStar detects evening star patterns
func (pd *PatternDetector) detectEveningStar(first, second, third *models.OHLCV) *CandlestickPattern {
	// First candle: bullish
	// Second candle: small body (doji/spinning top)
	// Third candle: bearish, closes below midpoint of first candle

	if first.Close <= first.Open || third.Close >= third.Open {
		return nil
	}

	firstBody := first.Close - first.Open
	secondBody := math.Abs(second.Close - second.Open)
	thirdBody := third.Open - third.Close

	firstMidpoint := (first.Open + first.Close) / 2

	// Check pattern conditions
	if secondBody < firstBody*0.5 && // Small middle candle
		third.Close < firstMidpoint && // Third closes below first's midpoint
		thirdBody > firstBody*0.3 { // Third has decent size

		return &CandlestickPattern{
			Name:        "Evening Star",
			Type:        "bearish",
			Strength:    75.0,
			Description: "Strong three-candle bearish reversal pattern",
		}
	}

	return nil
}

// GetStrongestPattern returns the pattern with highest strength
func GetStrongestPattern(patterns []*CandlestickPattern) *CandlestickPattern {
	if len(patterns) == 0 {
		return nil
	}

	strongest := patterns[0]
	for _, pattern := range patterns[1:] {
		if pattern.Strength > strongest.Strength {
			strongest = pattern
		}
	}

	return strongest
}
