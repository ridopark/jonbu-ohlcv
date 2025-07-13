package indicators

import (
	"sync"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-210: Performance-optimized caching

// CacheEntry represents a cached indicator calculation
type CacheEntry struct {
	TrendIndicators      *TrendIndicators      `json:"trend"`
	MomentumIndicators   *MomentumIndicators   `json:"momentum"`
	VolatilityIndicators *VolatilityIndicators `json:"volatility"`
	VolumeIndicators     *VolumeIndicators     `json:"volume"`
	CalculatedAt         time.Time             `json:"calculated_at"`
	CandleCount          int                   `json:"candle_count"`
}

// IndicatorCache provides caching for indicator calculations
type IndicatorCache struct {
	cache map[string]*CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewIndicatorCache creates a new indicator cache
func NewIndicatorCache(ttl time.Duration) *IndicatorCache {
	return &IndicatorCache{
		cache: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
}

// Get retrieves cached indicators for a symbol
func (c *IndicatorCache) Get(symbol string, candleCount int) *CacheEntry {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[symbol]
	if !exists {
		return nil
	}

	// Check if cache is expired
	if time.Since(entry.CalculatedAt) > c.ttl {
		return nil
	}

	// Check if candle count has changed significantly
	if entry.CandleCount != candleCount {
		return nil
	}

	return entry
}

// Set stores calculated indicators in cache
func (c *IndicatorCache) Set(symbol string, entry *CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry.CalculatedAt = time.Now()
	c.cache[symbol] = entry
}

// Clear removes a symbol from cache
func (c *IndicatorCache) Clear(symbol string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, symbol)
}

// CleanExpired removes expired cache entries
func (c *IndicatorCache) CleanExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for symbol, entry := range c.cache {
		if now.Sub(entry.CalculatedAt) > c.ttl {
			delete(c.cache, symbol)
		}
	}
}

// Size returns the number of cached entries
func (c *IndicatorCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// IndicatorSet represents a complete set of technical indicators
type IndicatorSet struct {
	Trend      *TrendIndicators      `json:"trend"`
	Momentum   *MomentumIndicators   `json:"momentum"`
	Volatility *VolatilityIndicators `json:"volatility"`
	Volume     *VolumeIndicators     `json:"volume"`
}

// CalculateAllIndicators computes all indicators with caching
func CalculateAllIndicators(symbol string, candles []*models.OHLCV, cache *IndicatorCache) *IndicatorSet {
	// Try to get from cache first
	if cache != nil {
		if cached := cache.Get(symbol, len(candles)); cached != nil {
			return &IndicatorSet{
				Trend:      cached.TrendIndicators,
				Momentum:   cached.MomentumIndicators,
				Volatility: cached.VolatilityIndicators,
				Volume:     cached.VolumeIndicators,
			}
		}
	}

	// Calculate indicators
	indicators := &IndicatorSet{
		Trend:      CalculateTrendIndicators(candles),
		Momentum:   CalculateMomentumIndicators(candles),
		Volatility: CalculateVolatilityIndicators(candles),
		Volume:     CalculateVolumeIndicators(candles),
	}

	// Store in cache
	if cache != nil {
		entry := &CacheEntry{
			TrendIndicators:      indicators.Trend,
			MomentumIndicators:   indicators.Momentum,
			VolatilityIndicators: indicators.Volatility,
			VolumeIndicators:     indicators.Volume,
			CandleCount:          len(candles),
		}
		cache.Set(symbol, entry)
	}

	return indicators
}

// OverallSignal provides a combined signal from all indicators
func (i *IndicatorSet) OverallSignal() string {
	bullishCount := 0
	bearishCount := 0

	// Trend signals
	if i.Trend.TrendDirection() == "bullish" {
		bullishCount++
	} else if i.Trend.TrendDirection() == "bearish" {
		bearishCount++
	}

	// Momentum signals
	momentum := i.Momentum.MomentumSignal()
	if momentum == "oversold" {
		bullishCount++
	} else if momentum == "overbought" {
		bearishCount++
	}

	// Volume confirmation
	if i.Volume.VolumeConfirmation() {
		if i.Volume.AccumulationSignal() == "accumulation" {
			bullishCount++
		} else if i.Volume.AccumulationSignal() == "distribution" {
			bearishCount++
		}
	}

	if bullishCount > bearishCount {
		return "bullish"
	} else if bearishCount > bullishCount {
		return "bearish"
	}

	return "neutral"
}

// Confidence returns confidence level (0-100) for the overall signal
func (i *IndicatorSet) Confidence() float64 {
	// Simple confidence calculation based on indicator alignment
	signals := []string{
		i.Trend.TrendDirection(),
		i.Momentum.MomentumSignal(),
		i.Volume.AccumulationSignal(),
	}

	overallSignal := i.OverallSignal()
	if overallSignal == "neutral" {
		return 50.0
	}

	agreement := 0
	for _, signal := range signals {
		if (overallSignal == "bullish" && (signal == "bullish" || signal == "oversold" || signal == "accumulation")) ||
			(overallSignal == "bearish" && (signal == "bearish" || signal == "overbought" || signal == "distribution")) {
			agreement++
		}
	}

	// Base confidence + agreement bonus
	confidence := 60.0 + (float64(agreement) * 10.0)
	if confidence > 95.0 {
		confidence = 95.0
	}

	return confidence
}
