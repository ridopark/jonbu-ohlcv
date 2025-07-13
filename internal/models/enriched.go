package models

import (
	"time"
)

// REQ-200 to REQ-205: Enriched candle with AI insights

// EnrichedCandle extends OHLCV with technical indicators and market context
type EnrichedCandle struct {
	// Base OHLCV data
	OHLCV *OHLCV `json:"ohlcv"`

	// Technical indicators
	Indicators *TechnicalIndicators `json:"indicators"`

	// Market analysis
	Analysis *MarketAnalysis `json:"analysis"`

	// Signal strength and confidence
	Signals *TradingSignals `json:"signals"`

	// Performance metadata
	Metadata *CandleMetadata `json:"metadata"`
}

// TechnicalIndicators contains all calculated indicators
type TechnicalIndicators struct {
	// Trend indicators
	SMA20 float64   `json:"sma_20"`
	SMA50 float64   `json:"sma_50"`
	EMA12 float64   `json:"ema_12"`
	EMA26 float64   `json:"ema_26"`
	MACD  *MACDData `json:"macd"`

	// Momentum indicators
	RSI        float64         `json:"rsi"`
	Stochastic *StochasticData `json:"stochastic"`
	WilliamsR  float64         `json:"williams_r"`

	// Volatility indicators
	BollingerBands *BollingerBandsData `json:"bollinger_bands"`
	ATR            float64             `json:"atr"`

	// Volume indicators
	VWAP      float64 `json:"vwap"`
	OBV       float64 `json:"obv"`
	VolumeMA  float64 `json:"volume_ma"`
	AccumDist float64 `json:"accum_dist"`

	// Trend analysis
	TrendDirection string  `json:"trend_direction"` // bullish, bearish, sideways
	TrendStrength  float64 `json:"trend_strength"`  // 0-100

	// Momentum analysis
	MomentumDirection string  `json:"momentum_direction"` // bullish, bearish, neutral
	MomentumStrength  float64 `json:"momentum_strength"`  // 0-100

	// Volatility analysis
	VolatilityLevel   string  `json:"volatility_level"`   // low, normal, high
	VolatilityPercent float64 `json:"volatility_percent"` // 0-100

	// Volume analysis
	VolumeConfirmation string  `json:"volume_confirmation"` // confirmed, weak, divergent
	RelativeVolume     float64 `json:"relative_volume"`     // vs average
}

// MarketAnalysis contains pattern and regime analysis
type MarketAnalysis struct {
	// Candlestick patterns
	CandlestickPatterns []string `json:"candlestick_patterns"`

	// Chart patterns
	ChartPatterns []ChartPatternResult `json:"chart_patterns"`

	// Market regime
	MarketRegime string `json:"market_regime"` // accumulation, markup, distribution, markdown

	// Support and resistance
	SupportResistance *SupportResistanceLevels `json:"support_resistance"`

	// Market context
	MarketPhase   string `json:"market_phase"` // opening, midday, closing
	SessionType   string `json:"session_type"` // regular, extended
	DayOfWeek     string `json:"day_of_week"`
	MarketHours   bool   `json:"market_hours"`
	VolumeProfile string `json:"volume_profile"` // low, normal, high, spike
}

// TradingSignals provides consolidated signal information
type TradingSignals struct {
	// Overall signal
	OverallSignal  string  `json:"overall_signal"`  // bullish, bearish, neutral
	SignalStrength float64 `json:"signal_strength"` // 0-100
	Confidence     float64 `json:"confidence"`      // 0-100

	// Component signals
	TrendSignal    string `json:"trend_signal"`    // bullish, bearish, neutral
	MomentumSignal string `json:"momentum_signal"` // bullish, bearish, neutral
	VolumeSignal   string `json:"volume_signal"`   // bullish, bearish, neutral

	// Entry/exit suggestions
	EntryLevel float64   `json:"entry_level,omitempty"`
	StopLoss   float64   `json:"stop_loss,omitempty"`
	TakeProfit []float64 `json:"take_profit,omitempty"`

	// Risk assessment
	RiskLevel  string  `json:"risk_level"` // low, medium, high
	Volatility float64 `json:"volatility"` // expected volatility

	// Pattern-based signals
	PatternSignals []PatternSignal `json:"pattern_signals,omitempty"`
}

// PatternSignal represents a signal from a specific pattern
type PatternSignal struct {
	PatternType string  `json:"pattern_type"`
	PatternName string  `json:"pattern_name"`
	Signal      string  `json:"signal"`     // bullish, bearish, neutral
	Confidence  float64 `json:"confidence"` // 0-100
	Target      float64 `json:"target,omitempty"`
	StopLoss    float64 `json:"stop_loss,omitempty"`
	TimeHorizon string  `json:"time_horizon"` // short, medium, long
}

// CandleMetadata contains performance and generation information
type CandleMetadata struct {
	// Generation timing
	GeneratedAt      time.Time `json:"generated_at"`
	ProcessingTimeMs float64   `json:"processing_time_ms"`

	// Data quality
	DataQuality       string  `json:"data_quality"`       // high, medium, low
	IndicatorCoverage float64 `json:"indicator_coverage"` // percentage of indicators calculated

	// Cache information
	CacheHits       int     `json:"cache_hits"`
	CacheMisses     int     `json:"cache_misses"`
	CacheEfficiency float64 `json:"cache_efficiency"`

	// Version information
	EngineVersion string `json:"engine_version"`
	ModelVersion  string `json:"model_version,omitempty"`

	// Debug information
	WarningsCount int      `json:"warnings_count"`
	Warnings      []string `json:"warnings,omitempty"`
}

// Supporting data structures

// MACDData contains MACD indicator values
type MACDData struct {
	Line      float64 `json:"line"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

// StochasticData contains Stochastic oscillator values
type StochasticData struct {
	K         float64 `json:"k"`
	D         float64 `json:"d"`
	Condition string  `json:"condition"` // oversold, neutral, overbought
}

// BollingerBandsData contains Bollinger Bands values
type BollingerBandsData struct {
	Upper     float64 `json:"upper"`
	Middle    float64 `json:"middle"`
	Lower     float64 `json:"lower"`
	Bandwidth float64 `json:"bandwidth"`
	Position  string  `json:"position"` // below_lower, middle, above_upper
}

// ChartPatternResult contains chart pattern detection results
type ChartPatternResult struct {
	Type       string  `json:"type"`       // triangle, breakout, head_shoulders, etc.
	Name       string  `json:"name"`       // descriptive name
	Confidence float64 `json:"confidence"` // 0-100
	Signal     string  `json:"signal"`     // bullish, bearish, neutral
	Target     float64 `json:"target,omitempty"`
	StopLoss   float64 `json:"stop_loss,omitempty"`
	Timeframe  string  `json:"timeframe"` // short, medium, long
	Status     string  `json:"status"`    // forming, confirmed, completed
}

// SupportResistanceLevels is imported from analysis package but redefined here for models
type SupportResistanceLevels struct {
	Support    []*SupportResistanceLevel `json:"support"`
	Resistance []*SupportResistanceLevel `json:"resistance"`
	Current    *CurrentLevelInfo         `json:"current"`
}

// SupportResistanceLevel represents a support or resistance level
type SupportResistanceLevel struct {
	Price      float64 `json:"price"`
	Type       string  `json:"type"`       // support, resistance
	Strength   float64 `json:"strength"`   // 0-100
	Touches    int     `json:"touches"`    // number of times price touched this level
	LastTouch  int     `json:"last_touch"` // periods ago
	Confidence float64 `json:"confidence"` // 0-100
}

// CurrentLevelInfo provides context about current price position
type CurrentLevelInfo struct {
	Price                float64                 `json:"price"`
	NearestSupport       *SupportResistanceLevel `json:"nearest_support"`
	NearestResistance    *SupportResistanceLevel `json:"nearest_resistance"`
	DistanceToSupport    float64                 `json:"distance_to_support"`    // percentage
	DistanceToResistance float64                 `json:"distance_to_resistance"` // percentage
	Position             string                  `json:"position"`               // near_support, near_resistance, middle
}

// EnrichmentOptions controls which enrichments to apply
type EnrichmentOptions struct {
	// Indicator categories
	TrendIndicators      bool `json:"trend_indicators"`
	MomentumIndicators   bool `json:"momentum_indicators"`
	VolatilityIndicators bool `json:"volatility_indicators"`
	VolumeIndicators     bool `json:"volume_indicators"`

	// Analysis categories
	CandlestickPatterns bool `json:"candlestick_patterns"`
	ChartPatterns       bool `json:"chart_patterns"`
	MarketRegime        bool `json:"market_regime"`
	SupportResistance   bool `json:"support_resistance"`

	// Signal generation
	TradingSignals bool `json:"trading_signals"`
	RiskAssessment bool `json:"risk_assessment"`

	// Performance options
	UseCache        bool `json:"use_cache"`
	IncludeMetadata bool `json:"include_metadata"`
	IncludeWarnings bool `json:"include_warnings"`
}

// DefaultEnrichmentOptions returns standard enrichment settings
func DefaultEnrichmentOptions() *EnrichmentOptions {
	return &EnrichmentOptions{
		TrendIndicators:      true,
		MomentumIndicators:   true,
		VolatilityIndicators: true,
		VolumeIndicators:     true,
		CandlestickPatterns:  true,
		ChartPatterns:        true,
		MarketRegime:         true,
		SupportResistance:    true,
		TradingSignals:       true,
		RiskAssessment:       true,
		UseCache:             true,
		IncludeMetadata:      true,
		IncludeWarnings:      false,
	}
}

// FastEnrichmentOptions returns performance-optimized settings
func FastEnrichmentOptions() *EnrichmentOptions {
	return &EnrichmentOptions{
		TrendIndicators:      true,
		MomentumIndicators:   true,
		VolatilityIndicators: false,
		VolumeIndicators:     true,
		CandlestickPatterns:  true,
		ChartPatterns:        false,
		MarketRegime:         false,
		SupportResistance:    false,
		TradingSignals:       true,
		RiskAssessment:       false,
		UseCache:             true,
		IncludeMetadata:      false,
		IncludeWarnings:      false,
	}
}

// ComprehensiveEnrichmentOptions returns all available enrichments
func ComprehensiveEnrichmentOptions() *EnrichmentOptions {
	return &EnrichmentOptions{
		TrendIndicators:      true,
		MomentumIndicators:   true,
		VolatilityIndicators: true,
		VolumeIndicators:     true,
		CandlestickPatterns:  true,
		ChartPatterns:        true,
		MarketRegime:         true,
		SupportResistance:    true,
		TradingSignals:       true,
		RiskAssessment:       true,
		UseCache:             true,
		IncludeMetadata:      true,
		IncludeWarnings:      true,
	}
}
