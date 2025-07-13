package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

func TestCandleEnrichmentEngine(t *testing.T) {
	// Create engine
	config := DefaultEnrichmentConfig()
	engine := NewCandleEnrichmentEngine(config)

	// Create test data
	history := generateTestCandles(50)
	current := &models.OHLCV{
		Timestamp: time.Now(),
		Open:      100.0,
		High:      105.0,
		Low:       99.0,
		Close:     103.0,
		Volume:    10000,
	}

	// Create enrichment options
	options := &models.EnrichmentOptions{
		TrendIndicators:      true,
		MomentumIndicators:   true,
		VolatilityIndicators: true,
		VolumeIndicators:     true,
		CandlestickPatterns:  true,
		ChartPatterns:        true,
		MarketRegime:         true,
		SupportResistance:    true,
		TradingSignals:       true,
		IncludeMetadata:      true,
	}

	// Test enrichment
	ctx := context.Background()
	enriched, err := engine.EnrichCandle(ctx, current, history, options)

	if err != nil {
		t.Fatalf("Failed to enrich candle: %v", err)
	}

	// Verify enriched candle
	if enriched == nil {
		t.Fatal("Enriched candle is nil")
	}

	if enriched.OHLCV != current {
		t.Error("OHLCV data mismatch")
	}

	// Verify indicators
	if enriched.Indicators == nil {
		t.Error("Technical indicators are nil")
	} else {
		if enriched.Indicators.SMA20 == 0 {
			t.Error("SMA20 not calculated")
		}
		if enriched.Indicators.RSI == 0 {
			t.Error("RSI not calculated")
		}
	}

	// Verify analysis
	if enriched.Analysis == nil {
		t.Error("Market analysis is nil")
	}

	// Verify signals
	if enriched.Signals == nil {
		t.Error("Trading signals are nil")
	} else {
		if enriched.Signals.OverallSignal == "" {
			t.Error("Overall signal not generated")
		}
		if enriched.Signals.Confidence == 0 {
			t.Error("Signal confidence not calculated")
		}
	}

	// Verify metadata
	if enriched.Metadata == nil {
		t.Error("Metadata is nil")
	} else {
		if enriched.Metadata.ProcessingTimeMs == 0 {
			t.Error("Processing time not recorded")
		}
		if enriched.Metadata.EngineVersion == "" {
			t.Error("Engine version not set")
		}
	}

	// Test performance requirement (< 1ms for REQ-205)
	if enriched.Metadata.ProcessingTimeMs > 1.0 {
		t.Logf("Warning: Processing time %f ms exceeds 1ms requirement",
			enriched.Metadata.ProcessingTimeMs)
	}

	t.Logf("Successfully enriched candle in %f ms", enriched.Metadata.ProcessingTimeMs)
}

func TestEnrichmentConfigDefaults(t *testing.T) {
	config := DefaultEnrichmentConfig()

	if config.MaxConcurrency <= 0 {
		t.Error("Invalid max concurrency")
	}

	if config.TimeoutMs <= 0 {
		t.Error("Invalid timeout")
	}

	if config.MinHistoryPeriods <= 0 {
		t.Error("Invalid min history periods")
	}
}

func TestInsufficientHistory(t *testing.T) {
	engine := NewCandleEnrichmentEngine(nil)

	// Test with insufficient history
	current := &models.OHLCV{
		Timestamp: time.Now(),
		Close:     100.0,
	}

	history := generateTestCandles(5) // Less than minimum required

	ctx := context.Background()
	_, err := engine.EnrichCandle(ctx, current, history, nil)

	if err == nil {
		t.Error("Expected error for insufficient history")
	}
}

// Helper function to generate test candles
func generateTestCandles(count int) []*models.OHLCV {
	candles := make([]*models.OHLCV, count)
	baseTime := time.Now().Add(-time.Duration(count) * time.Minute)
	basePrice := 100.0

	for i := 0; i < count; i++ {
		price := basePrice + float64(i%10) - 5.0 // Create some price variation

		candles[i] = &models.OHLCV{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      price,
			High:      price + 1.0,
			Low:       price - 1.0,
			Close:     price + 0.5,
			Volume:    int64(1000 + i*100),
		}
	}

	return candles
}
