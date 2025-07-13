package analysis

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-213: Market regime identification (Wyckoff methodology)

// RegimeAnalyzer identifies market regimes using Wyckoff methodology
type RegimeAnalyzer struct {
	lookbackPeriod int     // Period for analysis
	threshold      float64 // Threshold for regime changes
}

// NewRegimeAnalyzer creates a new regime analyzer
func NewRegimeAnalyzer() *RegimeAnalyzer {
	return &RegimeAnalyzer{
		lookbackPeriod: 50,  // 50 periods for analysis
		threshold:      0.1, // 10% threshold for significant moves
	}
}

// DetectRegime analyzes market regime using Wyckoff principles
func (ra *RegimeAnalyzer) DetectRegime(candles []*models.OHLCV) *MarketRegime {
	if len(candles) < ra.lookbackPeriod {
		return &MarketRegime{
			Phase:      "accumulation",
			Confidence: 50.0,
			Strength:   50.0,
		}
	}

	// Analyze price action
	priceAction := ra.analyzePriceAction(candles)

	// Analyze volume profile
	volumeProfile := ra.analyzeVolumeProfile(candles)

	// Analyze volatility patterns
	volatility := ra.analyzeVolatilityPattern(candles)

	// Determine regime based on Wyckoff phases
	phase := ra.determineWyckoffPhase(priceAction, volumeProfile, volatility)

	return &MarketRegime{
		Phase:         phase,
		Confidence:    ra.calculateConfidence(priceAction, volumeProfile, volatility),
		Strength:      ra.calculateStrength(priceAction, volumeProfile),
		PriceAction:   priceAction,
		VolumeProfile: volumeProfile,
		Volatility:    volatility,
	}
}

// MarketRegime represents the current market regime
type MarketRegime struct {
	Phase         string             `json:"phase"`      // accumulation, markup, distribution, markdown
	Confidence    float64            `json:"confidence"` // 0-100
	Strength      float64            `json:"strength"`   // 0-100
	PriceAction   *PriceActionData   `json:"price_action"`
	VolumeProfile *VolumeProfileData `json:"volume_profile"`
	Volatility    *VolatilityData    `json:"volatility"`
}

// PriceActionData contains price action analysis
type PriceActionData struct {
	Trend         string  `json:"trend"`          // up, down, sideways
	TrendStrength float64 `json:"trend_strength"` // 0-100
	Support       float64 `json:"support"`
	Resistance    float64 `json:"resistance"`
	Range         float64 `json:"range"`     // percentage range
	Breakouts     int     `json:"breakouts"` // number of recent breakouts
}

// VolumeProfileData contains volume analysis
type VolumeProfileData struct {
	RelativeVolume float64 `json:"relative_volume"` // current vs average
	VolumeMA       float64 `json:"volume_ma"`
	Distribution   string  `json:"distribution"` // accumulation, distribution, neutral
	Climax         bool    `json:"climax"`       // volume climax detected
	DryUp          bool    `json:"dry_up"`       // volume dry up detected
}

// VolatilityData contains volatility analysis
type VolatilityData struct {
	Level       string `json:"level"`       // low, normal, high
	Trending    bool   `json:"trending"`    // trending vs ranging market
	Expansion   bool   `json:"expansion"`   // volatility expansion
	Contraction bool   `json:"contraction"` // volatility contraction
}

// Price action analysis methods

func (ra *RegimeAnalyzer) analyzePriceAction(candles []*models.OHLCV) *PriceActionData {
	if len(candles) < 20 {
		return &PriceActionData{
			Trend:         "sideways",
			TrendStrength: 50.0,
		}
	}

	recent := candles[len(candles)-20:]

	// Calculate trend
	firstPrice := recent[0].Close
	lastPrice := recent[len(recent)-1].Close
	trendChange := (lastPrice - firstPrice) / firstPrice * 100

	trend := "sideways"
	trendStrength := math.Abs(trendChange) * 5 // Scale to 0-100

	if trendChange > 2 {
		trend = "up"
	} else if trendChange < -2 {
		trend = "down"
	}

	if trendStrength > 100 {
		trendStrength = 100
	}

	// Find support and resistance
	high := recent[0].High
	low := recent[0].Low

	for _, candle := range recent {
		if candle.High > high {
			high = candle.High
		}
		if candle.Low < low {
			low = candle.Low
		}
	}

	priceRange := (high - low) / low * 100

	// Count breakouts
	breakouts := 0
	for i := 1; i < len(recent); i++ {
		prevHigh := recent[i-1].High
		prevLow := recent[i-1].Low

		if recent[i].Close > prevHigh*1.02 || recent[i].Close < prevLow*0.98 {
			breakouts++
		}
	}

	return &PriceActionData{
		Trend:         trend,
		TrendStrength: trendStrength,
		Support:       low,
		Resistance:    high,
		Range:         priceRange,
		Breakouts:     breakouts,
	}
}

func (ra *RegimeAnalyzer) analyzeVolumeProfile(candles []*models.OHLCV) *VolumeProfileData {
	if len(candles) < 20 {
		return &VolumeProfileData{
			RelativeVolume: 1.0,
			Distribution:   "neutral",
		}
	}

	recent := candles[len(candles)-10:]
	historical := candles[len(candles)-30 : len(candles)-10]

	// Calculate average volumes
	recentAvgVolume := ra.calculateAverageVolume(recent)
	historicalAvgVolume := ra.calculateAverageVolume(historical)

	relativeVolume := recentAvgVolume / historicalAvgVolume

	// Determine distribution pattern
	distribution := "neutral"
	if relativeVolume > 1.5 {
		// High volume - could be distribution or accumulation
		if ra.isUpTrend(recent) {
			distribution = "distribution" // High volume on up move = distribution
		} else {
			distribution = "accumulation" // High volume on down move = accumulation
		}
	}

	// Detect volume climax (very high volume)
	maxVolume := ra.getMaxVolume(recent)
	climax := maxVolume > historicalAvgVolume*3

	// Detect volume dry up (very low volume)
	minVolume := ra.getMinVolume(recent)
	dryUp := minVolume < historicalAvgVolume*0.3

	return &VolumeProfileData{
		RelativeVolume: relativeVolume,
		VolumeMA:       historicalAvgVolume,
		Distribution:   distribution,
		Climax:         climax,
		DryUp:          dryUp,
	}
}

func (ra *RegimeAnalyzer) analyzeVolatilityPattern(candles []*models.OHLCV) *VolatilityData {
	if len(candles) < 20 {
		return &VolatilityData{
			Level: "normal",
		}
	}

	recent := candles[len(candles)-10:]

	// Calculate average true range for volatility
	atr := ra.calculateATR(recent)
	currentPrice := recent[len(recent)-1].Close
	volatilityPercent := (atr / currentPrice) * 100

	level := "normal"
	if volatilityPercent > 3 {
		level = "high"
	} else if volatilityPercent < 1 {
		level = "low"
	}

	// Determine if trending or ranging
	trendStrength := ra.calculateTrendStrength(recent)
	trending := trendStrength > 60

	// Check for volatility expansion/contraction
	olderATR := ra.calculateATR(candles[len(candles)-20 : len(candles)-10])
	expansion := atr > olderATR*1.5
	contraction := atr < olderATR*0.7

	return &VolatilityData{
		Level:       level,
		Trending:    trending,
		Expansion:   expansion,
		Contraction: contraction,
	}
}

func (ra *RegimeAnalyzer) determineWyckoffPhase(priceAction *PriceActionData, volumeProfile *VolumeProfileData, volatility *VolatilityData) string {
	// Wyckoff methodology phase determination

	// Accumulation phase characteristics:
	// - Sideways price action
	// - High volume on declines (accumulation)
	// - Low volatility (contraction)
	if priceAction.Trend == "sideways" &&
		volumeProfile.Distribution == "accumulation" &&
		volatility.Level == "low" {
		return "accumulation"
	}

	// Markup phase characteristics:
	// - Strong uptrend
	// - Increasing volume on advances
	// - Volatility expansion
	if priceAction.Trend == "up" &&
		priceAction.TrendStrength > 70 &&
		volatility.Expansion {
		return "markup"
	}

	// Distribution phase characteristics:
	// - Sideways price action at highs
	// - High volume on advances (distribution)
	// - Increased volatility
	if priceAction.Trend == "sideways" &&
		volumeProfile.Distribution == "distribution" &&
		volumeProfile.Climax {
		return "distribution"
	}

	// Markdown phase characteristics:
	// - Strong downtrend
	// - High volume on declines
	// - High volatility
	if priceAction.Trend == "down" &&
		priceAction.TrendStrength > 70 &&
		volatility.Level == "high" {
		return "markdown"
	}

	// Default to accumulation if unclear
	return "accumulation"
}

func (ra *RegimeAnalyzer) calculateConfidence(priceAction *PriceActionData, volumeProfile *VolumeProfileData, volatility *VolatilityData) float64 {
	confidence := 50.0 // Base confidence

	// Strong trends increase confidence
	confidence += priceAction.TrendStrength * 0.3

	// Volume confirmation increases confidence
	if volumeProfile.RelativeVolume > 1.2 {
		confidence += 15
	}

	// Clear volatility patterns increase confidence
	if volatility.Expansion || volatility.Contraction {
		confidence += 10
	}

	// Multiple breakouts reduce confidence (choppy market)
	if priceAction.Breakouts > 3 {
		confidence -= 20
	}

	if confidence > 95 {
		confidence = 95
	}
	if confidence < 20 {
		confidence = 20
	}

	return confidence
}

func (ra *RegimeAnalyzer) calculateStrength(priceAction *PriceActionData, volumeProfile *VolumeProfileData) float64 {
	strength := priceAction.TrendStrength

	// Volume confirmation adds strength
	if volumeProfile.RelativeVolume > 1.5 {
		strength += 20
	}

	// Volume climax adds strength
	if volumeProfile.Climax {
		strength += 15
	}

	if strength > 100 {
		strength = 100
	}

	return strength
}

// Helper methods

func (ra *RegimeAnalyzer) calculateAverageVolume(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	sum := int64(0)
	for _, candle := range candles {
		sum += candle.Volume
	}

	return float64(sum) / float64(len(candles))
}

func (ra *RegimeAnalyzer) isUpTrend(candles []*models.OHLCV) bool {
	if len(candles) < 2 {
		return false
	}

	first := candles[0].Close
	last := candles[len(candles)-1].Close

	return last > first
}

func (ra *RegimeAnalyzer) getMaxVolume(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	max := float64(candles[0].Volume)
	for _, candle := range candles {
		if float64(candle.Volume) > max {
			max = float64(candle.Volume)
		}
	}

	return max
}

func (ra *RegimeAnalyzer) getMinVolume(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	min := float64(candles[0].Volume)
	for _, candle := range candles {
		if float64(candle.Volume) < min {
			min = float64(candle.Volume)
		}
	}

	return min
}

func (ra *RegimeAnalyzer) calculateATR(candles []*models.OHLCV) float64 {
	if len(candles) < 2 {
		return 0
	}

	sum := 0.0
	count := 0

	for i := 1; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)

		trueRange := math.Max(tr1, math.Max(tr2, tr3))
		sum += trueRange
		count++
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}

func (ra *RegimeAnalyzer) calculateTrendStrength(candles []*models.OHLCV) float64 {
	if len(candles) < 2 {
		return 50
	}

	first := candles[0].Close
	last := candles[len(candles)-1].Close

	change := math.Abs((last - first) / first * 100)

	// Scale to 0-100
	strength := change * 10
	if strength > 100 {
		strength = 100
	}

	return strength
}
