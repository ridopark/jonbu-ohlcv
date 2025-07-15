package worker

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// REQ-031: 10k+ events/second per worker
// REQ-032: Sub-millisecond aggregation latency
// REQ-034: Buffered channels preventing blocking
// REQ-035: Memory usage monitoring and cleanup

// SymbolWorker handles aggregation for a specific symbol and timeframe
type SymbolWorker struct {
	Symbol    string
	Timeframe string

	// Channels
	Input  chan models.MarketEvent
	Output chan models.Candle

	// Aggregation state
	currentCandle    *models.Candle
	intervalDuration time.Duration

	// Mock mode data-driven aggregation
	useMockMode         bool
	dataCountInInterval int
	targetDataCount     int

	// Context and lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	mu              sync.RWMutex
	eventsProcessed int64
	candlesEmitted  int64

	logger zerolog.Logger
}

// WorkerConfig holds configuration for symbol workers
type WorkerConfig struct {
	Symbol      string
	Timeframe   string
	BufferSize  int
	LogLevel    string
	UseMockMode bool // Enable data-driven aggregation for mock testing
}

// NewSymbolWorker creates a new worker for a symbol-timeframe combination
func NewSymbolWorker(config WorkerConfig, logger zerolog.Logger) *SymbolWorker {
	ctx, cancel := context.WithCancel(context.Background())

	intervalDuration := parseTimeframeDuration(config.Timeframe)
	targetDataCount := getTargetDataCount(config.Timeframe, config.UseMockMode)

	return &SymbolWorker{
		Symbol:              config.Symbol,
		Timeframe:           config.Timeframe,
		Input:               make(chan models.MarketEvent, config.BufferSize),
		Output:              make(chan models.Candle, 100), // REQ-034: Buffered output
		intervalDuration:    intervalDuration,
		useMockMode:         config.UseMockMode,
		dataCountInInterval: 0,
		targetDataCount:     targetDataCount,
		ctx:                 ctx,
		cancel:              cancel,
		logger: logger.With().
			Str("component", "symbol_worker").
			Str("symbol", config.Symbol).
			Str("timeframe", config.Timeframe).
			Bool("mock_mode", config.UseMockMode).
			Logger(),
	}
}

// Start begins the worker's processing loop
func (w *SymbolWorker) Start() {
	w.logger.Info().Msg("Symbol worker started")
	go w.run()
}

// Stop gracefully shuts down the worker
func (w *SymbolWorker) Stop() {
	w.logger.Info().Msg("Stopping symbol worker")
	w.cancel()
	close(w.Input)
}

// run is the main worker processing loop
func (w *SymbolWorker) run() {
	defer func() {
		close(w.Output)
		w.logger.Info().
			Int64("events_processed", w.eventsProcessed).
			Int64("candles_emitted", w.candlesEmitted).
			Msg("Symbol worker stopped")
	}()

	// REQ-032: Sub-millisecond processing target
	var ticker *time.Ticker
	if w.useMockMode {
		// In mock mode, check less frequently since we're data-driven
		ticker = time.NewTicker(500 * time.Millisecond)
	} else {
		// In real mode, check for time-based interval completion
		ticker = time.NewTicker(100 * time.Millisecond)
	}
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			// Emit final candle if exists
			if w.currentCandle != nil {
				w.emitCandle()
			}
			return

		case event, ok := <-w.Input:
			if !ok {
				return
			}
			w.processEvent(event)

		case <-ticker.C:
			if !w.useMockMode {
				w.checkIntervalCompletion()
			}
		}
	}
}

// processEvent processes a single market event
func (w *SymbolWorker) processEvent(event models.MarketEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.eventsProcessed++

	// In mock mode, increment data count before interval check
	if w.useMockMode {
		w.dataCountInInterval++
	}

	// Check if this event belongs to a new interval
	if w.shouldStartNewInterval(event.Timestamp) {
		if w.currentCandle != nil {
			w.emitCandle()
		}
		w.startNewCandle(event)
	} else {
		w.updateCandle(event)
	}

	// Log every 1000 events for monitoring
	if w.eventsProcessed%1000 == 0 {
		w.logger.Debug().
			Int64("events", w.eventsProcessed).
			Int64("candles", w.candlesEmitted).
			Msg("Worker progress")
	}
}

// shouldStartNewInterval determines if we should start a new candle interval
func (w *SymbolWorker) shouldStartNewInterval(timestamp time.Time) bool {
	if w.currentCandle == nil {
		return true
	}

	if w.useMockMode {
		// In mock mode, start new interval when data count target is reached
		// This ensures we create separate candles with different timestamps
		return w.dataCountInInterval >= w.targetDataCount
	}

	// In real mode, use time-based boundaries
	intervalStart := w.getIntervalStart(timestamp)
	currentIntervalStart := w.getIntervalStart(w.currentCandle.Timestamp)

	return !intervalStart.Equal(currentIntervalStart)
}

// startNewCandle begins a new candle with the given event
func (w *SymbolWorker) startNewCandle(event models.MarketEvent) {
	var intervalStart time.Time

	if w.useMockMode {
		// In mock mode, use the actual event timestamp to avoid database conflicts
		// Since mock mode is data-driven, we don't need to align to interval boundaries
		intervalStart = event.Timestamp
	} else {
		// In real mode, use calculated interval boundaries
		intervalStart = w.getIntervalStart(event.Timestamp)
	}

	// Handle bar events with full OHLC data differently
	if event.Type == "bar" && event.Open != 0 && event.High != 0 && event.Low != 0 && event.Close != 0 {
		// Bar event with complete OHLC data - use it directly
		w.currentCandle = &models.Candle{
			Symbol:    w.Symbol,
			Timestamp: intervalStart,
			Open:      event.Open,
			High:      event.High,
			Low:       event.Low,
			Close:     event.Close,
			Volume:    event.Volume,
			Interval:  w.Timeframe,
		}
	} else {
		// Trade/quote event - build OHLC from price
		w.currentCandle = &models.Candle{
			Symbol:    w.Symbol,
			Timestamp: intervalStart,
			Open:      event.Price,
			High:      event.Price,
			Low:       event.Price,
			Close:     event.Price,
			Volume:    event.Volume,
			Interval:  w.Timeframe,
		}
	}

	// Reset data count for mock mode
	if w.useMockMode {
		w.dataCountInInterval = 1 // This event counts as the first
	}

	w.logger.Debug().
		Time("interval_start", intervalStart).
		Float64("open_price", w.currentCandle.Open).
		Bool("mock_mode", w.useMockMode).
		Msg("Started new candle")
}

// updateCandle updates the current candle with the new event
func (w *SymbolWorker) updateCandle(event models.MarketEvent) {
	if w.currentCandle == nil {
		w.startNewCandle(event)
		return
	}

	// Handle bar events with full OHLC data differently
	if event.Type == "bar" && event.Open != 0 && event.High != 0 && event.Low != 0 && event.Close != 0 {
		// Bar event with complete OHLC data - update with aggregated values
		w.currentCandle.High = math.Max(w.currentCandle.High, event.High)
		w.currentCandle.Low = math.Min(w.currentCandle.Low, event.Low)
		w.currentCandle.Close = event.Close
		w.currentCandle.Volume += event.Volume
	} else {
		// Trade/quote event - update OHLC values from price
		if event.Price > w.currentCandle.High {
			w.currentCandle.High = event.Price
		}
		if event.Price < w.currentCandle.Low {
			w.currentCandle.Low = event.Price
		}
		w.currentCandle.Close = event.Price
		w.currentCandle.Volume += event.Volume
	}
}

// checkIntervalCompletion checks if the current interval should be completed
func (w *SymbolWorker) checkIntervalCompletion() {
	if w.currentCandle == nil {
		return
	}

	now := time.Now()
	intervalEnd := w.currentCandle.Timestamp.Add(w.intervalDuration)

	// If current time has passed the interval end, emit the candle
	if now.After(intervalEnd) {
		w.mu.Lock()
		w.emitCandle()
		w.mu.Unlock()
	}
}

// emitCandle sends the current candle to the output channel
func (w *SymbolWorker) emitCandle() {
	if w.currentCandle == nil {
		return
	}

	candle := *w.currentCandle // Copy the candle
	w.currentCandle = nil
	w.candlesEmitted++

	// Reset data count for next interval in mock mode
	if w.useMockMode {
		w.dataCountInInterval = 0
	}

	select {
	case w.Output <- candle:
		w.logger.Debug().
			Time("timestamp", candle.Timestamp).
			Float64("open", candle.Open).
			Float64("high", candle.High).
			Float64("low", candle.Low).
			Float64("close", candle.Close).
			Int64("volume", candle.Volume).
			Bool("mock_mode", w.useMockMode).
			Msg("Emitted candle")
	default:
		// REQ-034: Don't block if output buffer is full
		w.logger.Warn().Msg("Output buffer full, dropping candle")
	}
}

// getIntervalStart calculates the interval start time for a given timestamp
func (w *SymbolWorker) getIntervalStart(timestamp time.Time) time.Time {
	switch w.Timeframe {
	case "1min":
		return timestamp.Truncate(time.Minute)
	case "5min":
		minutes := timestamp.Minute() / 5 * 5
		return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Hour(), minutes, 0, 0, timestamp.Location())
	case "15min":
		minutes := timestamp.Minute() / 15 * 15
		return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Hour(), minutes, 0, 0, timestamp.Location())
	case "30min":
		minutes := timestamp.Minute() / 30 * 30
		return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(),
			timestamp.Hour(), minutes, 0, 0, timestamp.Location())
	case "1hour":
		return timestamp.Truncate(time.Hour)
	case "1day":
		return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(),
			0, 0, 0, 0, timestamp.Location())
	default:
		return timestamp.Truncate(time.Minute)
	}
}

// parseTimeframeDuration converts timeframe string to duration
func parseTimeframeDuration(timeframe string) time.Duration {
	switch timeframe {
	case "1min":
		return time.Minute
	case "5min":
		return 5 * time.Minute
	case "15min":
		return 15 * time.Minute
	case "30min":
		return 30 * time.Minute
	case "1hour":
		return time.Hour
	case "1day":
		return 24 * time.Hour
	default:
		return time.Minute
	}
}

// getTargetDataCount calculates how many data points are needed for an interval in mock mode
func getTargetDataCount(timeframe string, useMockMode bool) int {
	if !useMockMode {
		return 0 // Not used in real mode
	}

	switch timeframe {
	case "1min":
		return 1 // 1 mock data point = 1 minute
	case "5min":
		return 5 // 5 mock data points = 5 minutes
	case "15min":
		return 15 // 15 mock data points = 15 minutes
	case "30min":
		return 30 // 30 mock data points = 30 minutes
	case "1hour":
		return 60 // 60 mock data points = 1 hour
	case "1day":
		return 1440 // 1440 mock data points = 1 day (24 * 60)
	default:
		return 1
	}
}

// GetMetrics returns worker performance metrics
func (w *SymbolWorker) GetMetrics() (eventsProcessed, candlesEmitted int64) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.eventsProcessed, w.candlesEmitted
}

// GetStatus returns worker status information
func (w *SymbolWorker) GetStatus() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	status := map[string]interface{}{
		"symbol":           w.Symbol,
		"timeframe":        w.Timeframe,
		"events_processed": w.eventsProcessed,
		"candles_emitted":  w.candlesEmitted,
		"active":           w.ctx.Err() == nil,
		"mock_mode":        w.useMockMode,
	}

	if w.useMockMode {
		status["data_count_in_interval"] = w.dataCountInInterval
		status["target_data_count"] = w.targetDataCount
	}

	if w.currentCandle != nil {
		status["current_candle"] = map[string]interface{}{
			"timestamp": w.currentCandle.Timestamp,
			"open":      w.currentCandle.Open,
			"high":      w.currentCandle.High,
			"low":       w.currentCandle.Low,
			"close":     w.currentCandle.Close,
			"volume":    w.currentCandle.Volume,
		}
	}

	return status
}
