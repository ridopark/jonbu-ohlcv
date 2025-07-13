package enrichment

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/analysis"
	"github.com/ridopark/jonbu-ohlcv/internal/indicators"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// REQ-200 to REQ-205: Enriched candle pipeline

// CandleEnrichmentEngine enriches OHLCV candles with AI insights
type CandleEnrichmentEngine struct {
	// Components
	indicatorCalculator  *indicators.IndicatorCache
	candlestickAnalyzer  *analysis.CandlestickAnalyzer
	chartPatternAnalyzer *analysis.ChartPatternAnalyzer
	regimeAnalyzer       *analysis.RegimeAnalyzer
	supportAnalyzer      *analysis.SupportResistanceDetector

	// Configuration
	config *EnrichmentConfig
	logger zerolog.Logger

	// Performance monitoring
	metrics *EnrichmentMetrics
	mu      sync.RWMutex
}

// EnrichmentConfig controls enrichment behavior
type EnrichmentConfig struct {
	// Performance settings
	MaxConcurrency  int  `json:"max_concurrency"`
	TimeoutMs       int  `json:"timeout_ms"`
	EnableProfiling bool `json:"enable_profiling"`

	// Data requirements
	MinHistoryPeriods int `json:"min_history_periods"`
	MaxHistoryPeriods int `json:"max_history_periods"`

	// Quality settings
	RequiredDataQuality string `json:"required_data_quality"` // high, medium, low
	EnableValidation    bool   `json:"enable_validation"`

	// Feature flags
	EnableAdvancedPatterns  bool `json:"enable_advanced_patterns"`
	EnableMarketRegime      bool `json:"enable_market_regime"`
	EnableSupportResistance bool `json:"enable_support_resistance"`

	// Cache settings
	CacheTTLMinutes int `json:"cache_ttl_minutes"`
	MaxCacheSize    int `json:"max_cache_size"`
}

// EnrichmentMetrics tracks performance statistics
type EnrichmentMetrics struct {
	// Timing metrics
	TotalEnrichments int64   `json:"total_enrichments"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	MaxLatencyMs     float64 `json:"max_latency_ms"`
	LatencyP95Ms     float64 `json:"latency_p95_ms"`

	// Quality metrics
	CacheHitRate float64 `json:"cache_hit_rate"`
	SuccessRate  float64 `json:"success_rate"`
	ErrorRate    float64 `json:"error_rate"`

	// Component metrics
	IndicatorLatencyMs float64 `json:"indicator_latency_ms"`
	AnalysisLatencyMs  float64 `json:"analysis_latency_ms"`
	SignalLatencyMs    float64 `json:"signal_latency_ms"`

	// Resource metrics
	MemoryUsageMB  float64 `json:"memory_usage_mb"`
	GoroutineCount int     `json:"goroutine_count"`

	// Error tracking
	LastError  string `json:"last_error,omitempty"`
	ErrorCount int64  `json:"error_count"`

	// Updated timestamp
	LastUpdated time.Time `json:"last_updated"`
}

// NewCandleEnrichmentEngine creates a new enrichment engine
func NewCandleEnrichmentEngine(config *EnrichmentConfig) *CandleEnrichmentEngine {
	if config == nil {
		config = DefaultEnrichmentConfig()
	}

	logger := zerolog.New(nil).With().
		Str("component", "enrichment_engine").
		Logger()

	engine := &CandleEnrichmentEngine{
		indicatorCalculator:  indicators.NewIndicatorCache(5 * time.Minute),
		candlestickAnalyzer:  analysis.NewCandlestickAnalyzer(),
		chartPatternAnalyzer: analysis.NewChartPatternAnalyzer(),
		regimeAnalyzer:       analysis.NewRegimeAnalyzer(),
		supportAnalyzer:      analysis.NewSupportResistanceDetector(),
		config:               config,
		logger:               logger,
		metrics:              &EnrichmentMetrics{},
	}

	return engine
}

// EnrichCandle enriches a single candle with AI insights
func (engine *CandleEnrichmentEngine) EnrichCandle(
	ctx context.Context,
	current *models.OHLCV,
	history []*models.OHLCV,
	options *models.EnrichmentOptions,
) (*models.EnrichedCandle, error) {

	startTime := time.Now()
	defer func() {
		engine.updateMetrics(time.Since(startTime))
	}()

	// Validate inputs
	if err := engine.validateInputs(current, history, options); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create context with timeout
	enrichCtx, cancel := context.WithTimeout(ctx, time.Duration(engine.config.TimeoutMs)*time.Millisecond)
	defer cancel()

	// Prepare enriched candle
	enriched := &models.EnrichedCandle{
		OHLCV: current,
		Metadata: &models.CandleMetadata{
			GeneratedAt:   time.Now(),
			EngineVersion: "1.0.0",
		},
	}

	// Technical indicators
	if options.TrendIndicators || options.MomentumIndicators ||
		options.VolatilityIndicators || options.VolumeIndicators {
		indicators, err := engine.calculateIndicators(enrichCtx, current, history, options)
		if err != nil {
			engine.logger.Warn().Err(err).Msg("Failed to calculate indicators")
		} else {
			enriched.Indicators = indicators
		}
	}

	// Market analysis
	if options.CandlestickPatterns || options.ChartPatterns ||
		options.MarketRegime || options.SupportResistance {
		analysis, err := engine.performAnalysis(enrichCtx, current, history, options)
		if err != nil {
			engine.logger.Warn().Err(err).Msg("Failed to perform analysis")
		} else {
			enriched.Analysis = analysis
		}
	}

	// Generate trading signals
	if options.TradingSignals {
		signals, err := engine.generateSignals(enrichCtx, enriched, options)
		if err != nil {
			engine.logger.Warn().Err(err).Msg("Failed to generate trading signals")
		} else {
			enriched.Signals = signals
		}
	}

	// Complete metadata
	if options.IncludeMetadata {
		engine.completeMetadata(enriched, startTime, options)
	}

	// Validate output
	if engine.config.EnableValidation {
		if err := engine.validateOutput(enriched); err != nil {
			return nil, fmt.Errorf("output validation failed: %w", err)
		}
	}

	engine.logger.Debug().
		Str("symbol", current.Symbol).
		Dur("processing_time", time.Since(startTime)).
		Float64("signal_strength", getSignalStrength(enriched)).
		Msg("Candle enrichment completed")

	return enriched, nil
}

// calculateIndicators computes technical indicators using the actual available methods
func (engine *CandleEnrichmentEngine) calculateIndicators(
	ctx context.Context,
	current *models.OHLCV,
	history []*models.OHLCV,
	options *models.EnrichmentOptions,
) (*models.TechnicalIndicators, error) {

	allCandles := append(history, current)

	// Calculate individual indicators
	indicators := &models.TechnicalIndicators{}

	// Use actual available methods from our indicator packages
	if options.TrendIndicators {
		// SMA calculations
		if len(allCandles) >= 20 {
			sma20 := engine.calculateSMA(allCandles, 20)
			indicators.SMA20 = sma20
		}

		if len(allCandles) >= 50 {
			sma50 := engine.calculateSMA(allCandles, 50)
			indicators.SMA50 = sma50
		}

		// EMA calculations
		if len(allCandles) >= 12 {
			ema12 := engine.calculateEMA(allCandles, 12)
			indicators.EMA12 = ema12
		}

		if len(allCandles) >= 26 {
			ema26 := engine.calculateEMA(allCandles, 26)
			indicators.EMA26 = ema26
		}

		// MACD
		if len(allCandles) >= 26 {
			macd := engine.calculateMACD(allCandles)
			indicators.MACD = macd
		}

		// Trend analysis
		indicators.TrendDirection = engine.determineTrendDirection(allCandles)
		indicators.TrendStrength = engine.calculateTrendStrength(allCandles)
	}

	if options.MomentumIndicators {
		// RSI
		if len(allCandles) >= 14 {
			rsi := engine.calculateRSI(allCandles, 14)
			indicators.RSI = rsi
		}

		// Stochastic
		if len(allCandles) >= 14 {
			stochastic := engine.calculateStochastic(allCandles, 14, 3)
			indicators.Stochastic = stochastic
		}

		// Williams %R
		if len(allCandles) >= 14 {
			williamsR := engine.calculateWilliamsR(allCandles, 14)
			indicators.WilliamsR = williamsR
		}

		// Momentum analysis
		indicators.MomentumDirection = engine.determineMomentumDirection(allCandles)
		indicators.MomentumStrength = engine.calculateMomentumStrength(allCandles)
	}

	if options.VolatilityIndicators {
		// Bollinger Bands
		if len(allCandles) >= 20 {
			bollingerBands := engine.calculateBollingerBands(allCandles, 20, 2.0)
			indicators.BollingerBands = bollingerBands
		}

		// ATR
		if len(allCandles) >= 14 {
			atr := engine.calculateATR(allCandles, 14)
			indicators.ATR = atr
		}

		// Volatility analysis
		indicators.VolatilityLevel = engine.determineVolatilityLevel(allCandles)
		indicators.VolatilityPercent = engine.calculateVolatilityPercent(allCandles)
	}

	if options.VolumeIndicators {
		// VWAP
		vwap := engine.calculateVWAP(allCandles)
		indicators.VWAP = vwap

		// OBV
		obv := engine.calculateOBV(allCandles)
		indicators.OBV = obv

		// Volume MA
		if len(allCandles) >= 20 {
			volumeMA := engine.calculateVolumeMA(allCandles, 20)
			indicators.VolumeMA = volumeMA
		}

		// Accumulation/Distribution
		accumDist := engine.calculateAccumulationDistribution(allCandles)
		indicators.AccumDist = accumDist

		// Volume analysis
		indicators.VolumeConfirmation = engine.determineVolumeConfirmation(allCandles)
		indicators.RelativeVolume = engine.calculateRelativeVolume(allCandles)
	}

	return indicators, nil
}

// performAnalysis conducts market context analysis
func (engine *CandleEnrichmentEngine) performAnalysis(
	ctx context.Context,
	current *models.OHLCV,
	history []*models.OHLCV,
	options *models.EnrichmentOptions,
) (*models.MarketAnalysis, error) {

	analysis := &models.MarketAnalysis{}

	// Candlestick patterns
	if options.CandlestickPatterns {
		recentCandles := getRecentCandles(append(history, current), 5)
		patterns := engine.candlestickAnalyzer.DetectPatterns(recentCandles)
		analysis.CandlestickPatterns = patterns
	}

	// Chart patterns
	if options.ChartPatterns && engine.config.EnableAdvancedPatterns {
		allCandles := append(history, current)
		chartPatterns := engine.chartPatternAnalyzer.DetectPatterns(allCandles)

		// Convert to enriched format
		analysis.ChartPatterns = make([]models.ChartPatternResult, len(chartPatterns))
		for i, pattern := range chartPatterns {
			analysis.ChartPatterns[i] = models.ChartPatternResult{
				Type:       pattern.Type,
				Name:       pattern.Name,
				Confidence: pattern.Confidence,
				Signal:     pattern.Signal,
				Target:     pattern.Target,
				StopLoss:   pattern.StopLoss,
				Timeframe:  pattern.Timeframe,
				Status:     pattern.Status,
			}
		}
	}

	// Market regime
	if options.MarketRegime && engine.config.EnableMarketRegime {
		allCandles := append(history, current)
		regime := engine.regimeAnalyzer.DetectRegime(allCandles)
		analysis.MarketRegime = regime.Phase
	}

	// Support and resistance
	if options.SupportResistance && engine.config.EnableSupportResistance {
		allCandles := append(history, current)
		srLevels := engine.supportAnalyzer.DetectLevels(allCandles)
		analysis.SupportResistance = convertSRLevels(srLevels)
	}

	// Market context
	analysis.MarketPhase = getMarketPhase(current.Timestamp)
	analysis.SessionType = getSessionType(current.Timestamp)
	analysis.DayOfWeek = current.Timestamp.Weekday().String()
	analysis.MarketHours = isMarketHours(current.Timestamp)
	analysis.VolumeProfile = getVolumeProfile(current.Volume, history)

	return analysis, nil
}

// generateSignals creates trading signals from enriched data
func (engine *CandleEnrichmentEngine) generateSignals(
	ctx context.Context,
	enriched *models.EnrichedCandle,
	options *models.EnrichmentOptions,
) (*models.TradingSignals, error) {

	signals := &models.TradingSignals{}

	// Component signals
	if enriched.Indicators != nil {
		signals.TrendSignal = enriched.Indicators.TrendDirection
		signals.MomentumSignal = enriched.Indicators.MomentumDirection
		signals.VolumeSignal = enriched.Indicators.VolumeConfirmation
	}

	// Overall signal calculation
	signalScore := 0.0
	signalCount := 0.0

	// Trend component (40% weight)
	if signals.TrendSignal == "bullish" {
		signalScore += 40
	} else if signals.TrendSignal == "bearish" {
		signalScore -= 40
	}
	signalCount += 40

	// Momentum component (35% weight)
	if signals.MomentumSignal == "bullish" {
		signalScore += 35
	} else if signals.MomentumSignal == "bearish" {
		signalScore -= 35
	}
	signalCount += 35

	// Volume confirmation (25% weight)
	if signals.VolumeSignal == "confirmed" {
		signalScore += 25
	} else if signals.VolumeSignal == "divergent" {
		signalScore -= 15
	}
	signalCount += 25

	// Normalize signal
	if signalCount > 0 {
		normalizedScore := signalScore / signalCount * 100

		if normalizedScore > 60 {
			signals.OverallSignal = "bullish"
		} else if normalizedScore < -60 {
			signals.OverallSignal = "bearish"
		} else {
			signals.OverallSignal = "neutral"
		}

		signals.SignalStrength = math.Abs(normalizedScore)
	}

	// Confidence calculation
	confidence := 50.0 // Base confidence

	if enriched.Indicators != nil {
		// Higher confidence with strong trend
		confidence += enriched.Indicators.TrendStrength * 0.3

		// Higher confidence with strong momentum
		confidence += enriched.Indicators.MomentumStrength * 0.2

		// Volume confirmation increases confidence
		if enriched.Indicators.VolumeConfirmation == "confirmed" {
			confidence += 15
		}
	}

	// Pattern confirmation
	if enriched.Analysis != nil && len(enriched.Analysis.CandlestickPatterns) > 0 {
		confidence += 10
	}

	if confidence > 95 {
		confidence = 95
	}

	signals.Confidence = confidence

	// Risk assessment
	if options.RiskAssessment {
		signals.RiskLevel = calculateRiskLevel(enriched)
		if enriched.Indicators != nil {
			signals.Volatility = enriched.Indicators.VolatilityPercent
		}
	}

	// Pattern-based signals
	if enriched.Analysis != nil {
		signals.PatternSignals = generatePatternSignals(enriched.Analysis)
	}

	return signals, nil
}

// Helper functions

func DefaultEnrichmentConfig() *EnrichmentConfig {
	return &EnrichmentConfig{
		MaxConcurrency:          4,
		TimeoutMs:               1000,
		EnableProfiling:         false,
		MinHistoryPeriods:       20,
		MaxHistoryPeriods:       200,
		RequiredDataQuality:     "medium",
		EnableValidation:        true,
		EnableAdvancedPatterns:  true,
		EnableMarketRegime:      true,
		EnableSupportResistance: true,
		CacheTTLMinutes:         5,
		MaxCacheSize:            1000,
	}
}

func (engine *CandleEnrichmentEngine) validateInputs(current *models.OHLCV, history []*models.OHLCV, options *models.EnrichmentOptions) error {
	if current == nil {
		return fmt.Errorf("current candle is required")
	}

	if len(history) < engine.config.MinHistoryPeriods {
		return fmt.Errorf("insufficient history: need %d periods, got %d",
			engine.config.MinHistoryPeriods, len(history))
	}

	if options == nil {
		return fmt.Errorf("enrichment options are required")
	}

	return nil
}

func (engine *CandleEnrichmentEngine) validateOutput(enriched *models.EnrichedCandle) error {
	if enriched.OHLCV == nil {
		return fmt.Errorf("base OHLCV data is missing")
	}

	return nil
}

func (engine *CandleEnrichmentEngine) updateMetrics(duration time.Duration) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	engine.metrics.TotalEnrichments++
	latencyMs := float64(duration.Nanoseconds()) / 1e6

	// Update average latency
	if engine.metrics.AverageLatencyMs == 0 {
		engine.metrics.AverageLatencyMs = latencyMs
	} else {
		engine.metrics.AverageLatencyMs = (engine.metrics.AverageLatencyMs + latencyMs) / 2
	}

	// Update max latency
	if latencyMs > engine.metrics.MaxLatencyMs {
		engine.metrics.MaxLatencyMs = latencyMs
	}

	engine.metrics.LastUpdated = time.Now()
}

func (engine *CandleEnrichmentEngine) completeMetadata(enriched *models.EnrichedCandle, startTime time.Time, options *models.EnrichmentOptions) {
	enriched.Metadata.ProcessingTimeMs = float64(time.Since(startTime).Nanoseconds()) / 1e6
	enriched.Metadata.DataQuality = "high"
	enriched.Metadata.IndicatorCoverage = calculateIndicatorCoverage(enriched.Indicators)

	// Cache statistics
	engine.mu.RLock()
	enriched.Metadata.CacheEfficiency = engine.metrics.CacheHitRate
	engine.mu.RUnlock()
}

// Placeholder implementations for indicator calculations
// These would call the actual indicator functions from our indicators package

func (engine *CandleEnrichmentEngine) calculateSMA(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}

	sum := 0.0
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Close
	}

	return sum / float64(period)
}

func (engine *CandleEnrichmentEngine) calculateEMA(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}

	// Simplified EMA calculation
	multiplier := 2.0 / (float64(period) + 1.0)
	ema := candles[0].Close

	for i := 1; i < len(candles); i++ {
		ema = (candles[i].Close * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

func (engine *CandleEnrichmentEngine) calculateMACD(candles []*models.OHLCV) *models.MACDData {
	if len(candles) < 26 {
		return &models.MACDData{}
	}

	ema12 := engine.calculateEMA(candles, 12)
	ema26 := engine.calculateEMA(candles, 26)
	macdLine := ema12 - ema26

	return &models.MACDData{
		Line:      macdLine,
		Signal:    macdLine * 0.9, // Simplified signal line
		Histogram: macdLine * 0.1, // Simplified histogram
	}
}

func (engine *CandleEnrichmentEngine) calculateRSI(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period+1 {
		return 50 // Neutral RSI
	}

	gains := 0.0
	losses := 0.0

	for i := len(candles) - period; i < len(candles); i++ {
		if i == 0 {
			continue
		}

		change := candles[i].Close - candles[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
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

func (engine *CandleEnrichmentEngine) calculateStochastic(candles []*models.OHLCV, kPeriod, dPeriod int) *models.StochasticData {
	if len(candles) < kPeriod {
		return &models.StochasticData{K: 50, D: 50, Condition: "neutral"}
	}

	// Find highest high and lowest low in the period
	highestHigh := candles[len(candles)-kPeriod].High
	lowestLow := candles[len(candles)-kPeriod].Low

	for i := len(candles) - kPeriod + 1; i < len(candles); i++ {
		if candles[i].High > highestHigh {
			highestHigh = candles[i].High
		}
		if candles[i].Low < lowestLow {
			lowestLow = candles[i].Low
		}
	}

	currentClose := candles[len(candles)-1].Close
	k := ((currentClose - lowestLow) / (highestHigh - lowestLow)) * 100
	d := k * 0.9 // Simplified D calculation

	condition := "neutral"
	if k < 20 {
		condition = "oversold"
	} else if k > 80 {
		condition = "overbought"
	}

	return &models.StochasticData{
		K:         k,
		D:         d,
		Condition: condition,
	}
}

func (engine *CandleEnrichmentEngine) calculateWilliamsR(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		return -50 // Neutral Williams %R
	}

	// Similar to Stochastic but inverted
	stoch := engine.calculateStochastic(candles, period, 3)
	return stoch.K - 100
}

func (engine *CandleEnrichmentEngine) calculateBollingerBands(candles []*models.OHLCV, period int, stdDev float64) *models.BollingerBandsData {
	if len(candles) < period {
		currentPrice := candles[len(candles)-1].Close
		return &models.BollingerBandsData{
			Upper:     currentPrice * 1.02,
			Middle:    currentPrice,
			Lower:     currentPrice * 0.98,
			Bandwidth: 4.0,
			Position:  "middle",
		}
	}

	sma := engine.calculateSMA(candles, period)

	// Calculate standard deviation
	variance := 0.0
	for i := len(candles) - period; i < len(candles); i++ {
		variance += math.Pow(candles[i].Close-sma, 2)
	}
	variance /= float64(period)
	standardDeviation := math.Sqrt(variance)

	upper := sma + (stdDev * standardDeviation)
	lower := sma - (stdDev * standardDeviation)
	bandwidth := ((upper - lower) / sma) * 100

	currentPrice := candles[len(candles)-1].Close
	position := "middle"
	if currentPrice < lower {
		position = "below_lower"
	} else if currentPrice > upper {
		position = "above_upper"
	}

	return &models.BollingerBandsData{
		Upper:     upper,
		Middle:    sma,
		Lower:     lower,
		Bandwidth: bandwidth,
		Position:  position,
	}
}

func (engine *CandleEnrichmentEngine) calculateATR(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	trueRanges := make([]float64, 0, period)

	for i := len(candles) - period; i < len(candles); i++ {
		if i == 0 {
			continue
		}

		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)

		trueRange := math.Max(tr1, math.Max(tr2, tr3))
		trueRanges = append(trueRanges, trueRange)
	}

	sum := 0.0
	for _, tr := range trueRanges {
		sum += tr
	}

	return sum / float64(len(trueRanges))
}

func (engine *CandleEnrichmentEngine) calculateVWAP(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	totalPriceVolume := 0.0
	totalVolume := int64(0)

	for _, candle := range candles {
		typicalPrice := (candle.High + candle.Low + candle.Close) / 3
		totalPriceVolume += typicalPrice * float64(candle.Volume)
		totalVolume += candle.Volume
	}

	if totalVolume == 0 {
		return candles[len(candles)-1].Close
	}

	return totalPriceVolume / float64(totalVolume)
}

func (engine *CandleEnrichmentEngine) calculateOBV(candles []*models.OHLCV) float64 {
	if len(candles) < 2 {
		return 0
	}

	obv := 0.0

	for i := 1; i < len(candles); i++ {
		if candles[i].Close > candles[i-1].Close {
			obv += float64(candles[i].Volume)
		} else if candles[i].Close < candles[i-1].Close {
			obv -= float64(candles[i].Volume)
		}
		// If close is equal to previous close, OBV remains unchanged
	}

	return obv
}

func (engine *CandleEnrichmentEngine) calculateVolumeMA(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}

	sum := int64(0)
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Volume
	}

	return float64(sum) / float64(period)
}

func (engine *CandleEnrichmentEngine) calculateAccumulationDistribution(candles []*models.OHLCV) float64 {
	if len(candles) == 0 {
		return 0
	}

	ad := 0.0

	for _, candle := range candles {
		if candle.High != candle.Low {
			moneyFlowMultiplier := ((candle.Close - candle.Low) - (candle.High - candle.Close)) / (candle.High - candle.Low)
			moneyFlowVolume := moneyFlowMultiplier * float64(candle.Volume)
			ad += moneyFlowVolume
		}
	}

	return ad
}

// Analysis helper functions

func (engine *CandleEnrichmentEngine) determineTrendDirection(candles []*models.OHLCV) string {
	if len(candles) < 20 {
		return "neutral"
	}

	sma20 := engine.calculateSMA(candles, 20)
	sma50 := engine.calculateSMA(candles, 50)
	currentPrice := candles[len(candles)-1].Close

	if currentPrice > sma20 && sma20 > sma50 {
		return "bullish"
	} else if currentPrice < sma20 && sma20 < sma50 {
		return "bearish"
	}

	return "neutral"
}

func (engine *CandleEnrichmentEngine) calculateTrendStrength(candles []*models.OHLCV) float64 {
	if len(candles) < 20 {
		return 50
	}

	// Calculate slope of moving average to determine trend strength
	sma := engine.calculateSMA(candles, 20)
	olderSMA := engine.calculateSMA(candles[:len(candles)-10], 20)

	slope := (sma - olderSMA) / olderSMA
	strength := math.Abs(slope) * 1000 // Scale to 0-100 range

	if strength > 100 {
		strength = 100
	}

	return strength
}

func (engine *CandleEnrichmentEngine) determineMomentumDirection(candles []*models.OHLCV) string {
	if len(candles) < 14 {
		return "neutral"
	}

	rsi := engine.calculateRSI(candles, 14)

	if rsi > 60 {
		return "bullish"
	} else if rsi < 40 {
		return "bearish"
	}

	return "neutral"
}

func (engine *CandleEnrichmentEngine) calculateMomentumStrength(candles []*models.OHLCV) float64 {
	if len(candles) < 14 {
		return 50
	}

	rsi := engine.calculateRSI(candles, 14)

	// Convert RSI to momentum strength (distance from 50)
	return math.Abs(rsi-50) * 2
}

func (engine *CandleEnrichmentEngine) determineVolatilityLevel(candles []*models.OHLCV) string {
	if len(candles) < 14 {
		return "normal"
	}

	atr := engine.calculateATR(candles, 14)
	currentPrice := candles[len(candles)-1].Close

	atrPercent := (atr / currentPrice) * 100

	if atrPercent > 3.0 {
		return "high"
	} else if atrPercent < 1.0 {
		return "low"
	}

	return "normal"
}

func (engine *CandleEnrichmentEngine) calculateVolatilityPercent(candles []*models.OHLCV) float64 {
	if len(candles) < 14 {
		return 0
	}

	atr := engine.calculateATR(candles, 14)
	currentPrice := candles[len(candles)-1].Close

	return (atr / currentPrice) * 100
}

func (engine *CandleEnrichmentEngine) determineVolumeConfirmation(candles []*models.OHLCV) string {
	if len(candles) < 20 {
		return "neutral"
	}

	currentVolume := float64(candles[len(candles)-1].Volume)
	avgVolume := engine.calculateVolumeMA(candles[:len(candles)-1], 20)

	if currentVolume > avgVolume*1.5 {
		return "confirmed"
	} else if currentVolume < avgVolume*0.5 {
		return "weak"
	}

	return "neutral"
}

func (engine *CandleEnrichmentEngine) calculateRelativeVolume(candles []*models.OHLCV) float64 {
	if len(candles) < 20 {
		return 1.0
	}

	currentVolume := float64(candles[len(candles)-1].Volume)
	avgVolume := engine.calculateVolumeMA(candles[:len(candles)-1], 20)

	if avgVolume == 0 {
		return 1.0
	}

	return currentVolume / avgVolume
}

// Additional helper functions

func getSignalStrength(enriched *models.EnrichedCandle) float64 {
	if enriched.Signals == nil {
		return 0
	}
	return enriched.Signals.SignalStrength
}

func getRecentCandles(candles []*models.OHLCV, count int) []*models.OHLCV {
	if len(candles) <= count {
		return candles
	}
	return candles[len(candles)-count:]
}

func convertSRLevels(levels *analysis.SupportResistanceLevels) *models.SupportResistanceLevels {
	if levels == nil {
		return &models.SupportResistanceLevels{}
	}

	modelLevels := &models.SupportResistanceLevels{
		Support:    make([]*models.SupportResistanceLevel, len(levels.Support)),
		Resistance: make([]*models.SupportResistanceLevel, len(levels.Resistance)),
	}

	// Convert support levels
	for i, level := range levels.Support {
		modelLevels.Support[i] = &models.SupportResistanceLevel{
			Price:      level.Price,
			Type:       level.Type,
			Strength:   level.Strength,
			Touches:    level.Touches,
			LastTouch:  level.LastTouch,
			Confidence: level.Confidence,
		}
	}

	// Convert resistance levels
	for i, level := range levels.Resistance {
		modelLevels.Resistance[i] = &models.SupportResistanceLevel{
			Price:      level.Price,
			Type:       level.Type,
			Strength:   level.Strength,
			Touches:    level.Touches,
			LastTouch:  level.LastTouch,
			Confidence: level.Confidence,
		}
	}

	// Convert current level info
	if levels.Current != nil {
		modelLevels.Current = &models.CurrentLevelInfo{
			Price:                levels.Current.Price,
			DistanceToSupport:    levels.Current.DistanceToSupport,
			DistanceToResistance: levels.Current.DistanceToResistance,
			Position:             levels.Current.Position,
		}

		if levels.Current.NearestSupport != nil {
			modelLevels.Current.NearestSupport = &models.SupportResistanceLevel{
				Price:      levels.Current.NearestSupport.Price,
				Type:       levels.Current.NearestSupport.Type,
				Strength:   levels.Current.NearestSupport.Strength,
				Touches:    levels.Current.NearestSupport.Touches,
				LastTouch:  levels.Current.NearestSupport.LastTouch,
				Confidence: levels.Current.NearestSupport.Confidence,
			}
		}

		if levels.Current.NearestResistance != nil {
			modelLevels.Current.NearestResistance = &models.SupportResistanceLevel{
				Price:      levels.Current.NearestResistance.Price,
				Type:       levels.Current.NearestResistance.Type,
				Strength:   levels.Current.NearestResistance.Strength,
				Touches:    levels.Current.NearestResistance.Touches,
				LastTouch:  levels.Current.NearestResistance.LastTouch,
				Confidence: levels.Current.NearestResistance.Confidence,
			}
		}
	}

	return modelLevels
}

func getMarketPhase(timestamp time.Time) string {
	hour := timestamp.Hour()
	if hour < 10 {
		return "opening"
	} else if hour > 15 {
		return "closing"
	}
	return "midday"
}

func getSessionType(timestamp time.Time) string {
	hour := timestamp.Hour()
	if hour >= 9 && hour <= 16 {
		return "regular"
	}
	return "extended"
}

func isMarketHours(timestamp time.Time) bool {
	weekday := timestamp.Weekday()
	hour := timestamp.Hour()

	// Market is closed on weekends
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// Regular market hours: 9:30 AM - 4:00 PM ET
	return hour >= 9 && hour <= 16
}

func getVolumeProfile(currentVolume int64, history []*models.OHLCV) string {
	if len(history) == 0 {
		return "normal"
	}

	// Calculate average volume
	sum := int64(0)
	for _, candle := range history {
		sum += candle.Volume
	}
	avgVolume := float64(sum) / float64(len(history))

	ratio := float64(currentVolume) / avgVolume

	if ratio > 2.0 {
		return "spike"
	} else if ratio > 1.5 {
		return "high"
	} else if ratio < 0.5 {
		return "low"
	}
	return "normal"
}

func calculateRiskLevel(enriched *models.EnrichedCandle) string {
	risk := 0.0

	if enriched.Indicators != nil {
		// Higher volatility = higher risk
		risk += enriched.Indicators.VolatilityPercent * 0.01

		// Low volume confirmation = higher risk
		if enriched.Indicators.VolumeConfirmation == "weak" {
			risk += 0.3
		} else if enriched.Indicators.VolumeConfirmation == "divergent" {
			risk += 0.5
		}
	}

	if risk > 0.7 {
		return "high"
	} else if risk > 0.4 {
		return "medium"
	}
	return "low"
}

func generatePatternSignals(analysis *models.MarketAnalysis) []models.PatternSignal {
	signals := make([]models.PatternSignal, 0)

	// Convert candlestick patterns to signals
	for _, pattern := range analysis.CandlestickPatterns {
		signal := models.PatternSignal{
			PatternType: "candlestick",
			PatternName: pattern,
			Signal:      inferPatternSignal(pattern),
			Confidence:  75.0,
			TimeHorizon: "short",
		}
		signals = append(signals, signal)
	}

	// Convert chart patterns to signals
	for _, pattern := range analysis.ChartPatterns {
		signal := models.PatternSignal{
			PatternType: "chart",
			PatternName: pattern.Name,
			Signal:      pattern.Signal,
			Target:      pattern.Target,
			StopLoss:    pattern.StopLoss,
			Confidence:  pattern.Confidence,
			TimeHorizon: pattern.Timeframe,
		}
		signals = append(signals, signal)
	}

	return signals
}

func inferPatternSignal(patternName string) string {
	bullishPatterns := []string{"hammer", "doji", "engulfing_bullish", "morning_star"}
	bearishPatterns := []string{"shooting_star", "engulfing_bearish", "evening_star"}

	for _, pattern := range bullishPatterns {
		if patternName == pattern {
			return "bullish"
		}
	}

	for _, pattern := range bearishPatterns {
		if patternName == pattern {
			return "bearish"
		}
	}

	return "neutral"
}

func calculateIndicatorCoverage(indicators *models.TechnicalIndicators) float64 {
	if indicators == nil {
		return 0
	}

	total := 15.0 // Total number of indicator fields
	calculated := 0.0

	if indicators.SMA20 != 0 {
		calculated++
	}
	if indicators.SMA50 != 0 {
		calculated++
	}
	if indicators.EMA12 != 0 {
		calculated++
	}
	if indicators.EMA26 != 0 {
		calculated++
	}
	if indicators.MACD != nil {
		calculated++
	}
	if indicators.RSI != 0 {
		calculated++
	}
	if indicators.Stochastic != nil {
		calculated++
	}
	if indicators.WilliamsR != 0 {
		calculated++
	}
	if indicators.BollingerBands != nil {
		calculated++
	}
	if indicators.ATR != 0 {
		calculated++
	}
	if indicators.VWAP != 0 {
		calculated++
	}
	if indicators.OBV != 0 {
		calculated++
	}
	if indicators.VolumeMA != 0 {
		calculated++
	}
	if indicators.AccumDist != 0 {
		calculated++
	}
	if indicators.TrendDirection != "" {
		calculated++
	}

	return (calculated / total) * 100
}
