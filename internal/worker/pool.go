package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// REQ-031: 10k+ events/second processing capacity
// REQ-032: Sub-millisecond aggregation latency
// REQ-033: Dynamic worker scaling
// REQ-034: Buffered channels preventing blocking
// REQ-035: Memory usage monitoring and cleanup

// Pool manages multiple symbol workers for parallel processing
type Pool struct {
	// Worker management
	workers   map[string]*SymbolWorker // key: symbol:timeframe
	workersMu sync.RWMutex

	// Configuration
	config PoolConfig

	// Input channels
	eventInput chan models.MarketEvent

	// Output aggregation
	candleOutput chan models.Candle

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	totalEvents  int64
	totalCandles int64
	metricsMu    sync.RWMutex

	logger zerolog.Logger
}

// PoolConfig holds configuration for the worker pool
type PoolConfig struct {
	MaxWorkers          int
	EventBufferSize     int
	CandleBufferSize    int
	WorkerBufferSize    int
	HealthCheckInterval time.Duration
	MetricsInterval     time.Duration
	UseMockMode         bool // Enable data-driven aggregation for mock testing
}

// DefaultPoolConfig returns a default configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxWorkers:          100,
		EventBufferSize:     10000, // REQ-034: Large buffer for high throughput
		CandleBufferSize:    5000,
		WorkerBufferSize:    1000,
		HealthCheckInterval: 30 * time.Second,
		MetricsInterval:     60 * time.Second,
	}
}

// NewPool creates a new worker pool
func NewPool(config PoolConfig, logger zerolog.Logger) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workers:      make(map[string]*SymbolWorker),
		config:       config,
		eventInput:   make(chan models.MarketEvent, config.EventBufferSize),
		candleOutput: make(chan models.Candle, config.CandleBufferSize),
		ctx:          ctx,
		cancel:       cancel,
		logger: logger.With().
			Str("component", "worker_pool").
			Logger(),
	}
}

// Start begins the worker pool operations
func (p *Pool) Start() {
	p.logger.Info().
		Int("max_workers", p.config.MaxWorkers).
		Int("event_buffer", p.config.EventBufferSize).
		Msg("Starting worker pool")

	// Start event dispatcher
	p.wg.Add(1)
	go p.dispatchEvents()

	// Start metrics collector
	p.wg.Add(1)
	go p.collectMetrics()

	// Start health checker
	p.wg.Add(1)
	go p.healthCheck()
}

// Stop gracefully shuts down the worker pool
func (p *Pool) Stop() {
	p.logger.Info().Msg("Stopping worker pool")

	p.cancel()
	close(p.eventInput)

	// Stop all workers
	p.workersMu.Lock()
	for _, worker := range p.workers {
		worker.Stop()
	}
	p.workersMu.Unlock()

	// Wait for all goroutines to finish
	p.wg.Wait()
	close(p.candleOutput)

	p.logger.Info().Msg("Worker pool stopped")
}

// AddSymbol adds a new symbol-timeframe combination to the pool
func (p *Pool) AddSymbol(symbol, timeframe string) error {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	key := fmt.Sprintf("%s:%s", symbol, timeframe)

	// Check if worker already exists
	if _, exists := p.workers[key]; exists {
		return fmt.Errorf("worker for %s already exists", key)
	}

	// Check worker limit
	if len(p.workers) >= p.config.MaxWorkers {
		return fmt.Errorf("maximum worker limit reached (%d)", p.config.MaxWorkers)
	}

	// Create new worker
	workerConfig := WorkerConfig{
		Symbol:      symbol,
		Timeframe:   timeframe,
		BufferSize:  p.config.WorkerBufferSize,
		UseMockMode: p.config.UseMockMode,
	}

	worker := NewSymbolWorker(workerConfig, p.logger)
	p.workers[key] = worker

	// Start the worker
	worker.Start()

	p.logger.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Int("total_workers", len(p.workers)).
		Msg("Added new symbol worker")

	return nil
}

// RemoveSymbol removes a symbol-timeframe worker from the pool
func (p *Pool) RemoveSymbol(symbol, timeframe string) error {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	key := fmt.Sprintf("%s:%s", symbol, timeframe)

	worker, exists := p.workers[key]
	if !exists {
		return fmt.Errorf("worker for %s not found", key)
	}

	// Stop the worker
	worker.Stop()
	delete(p.workers, key)

	p.logger.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Int("total_workers", len(p.workers)).
		Msg("Removed symbol worker")

	return nil
}

// ProcessEvent sends an event to the appropriate worker
func (p *Pool) ProcessEvent(event models.MarketEvent) {
	select {
	case p.eventInput <- event:
		// Event queued successfully
	default:
		// REQ-034: Handle backpressure by dropping events
		p.logger.Warn().
			Str("symbol", event.Symbol).
			Msg("Event input buffer full, dropping event")
	}
}

// GetCandleOutput returns the channel for consuming aggregated candles
func (p *Pool) GetCandleOutput() <-chan models.Candle {
	return p.candleOutput
}

// dispatchEvents distributes events to appropriate workers
func (p *Pool) dispatchEvents() {
	defer p.wg.Done()

	p.logger.Info().Msg("Event dispatcher started")

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info().Msg("Event dispatcher stopping")
			return

		case event, ok := <-p.eventInput:
			if !ok {
				p.logger.Info().Msg("Event input closed, dispatcher stopping")
				return
			}

			p.dispatchToWorkers(event)
		}
	}
}

// dispatchToWorkers sends an event to all relevant workers
func (p *Pool) dispatchToWorkers(event models.MarketEvent) {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()

	// REQ-031: High-performance event distribution
	for key, worker := range p.workers {
		// Check if this worker handles this symbol
		if worker.Symbol == event.Symbol {
			select {
			case worker.Input <- event:
				// Event sent successfully
			default:
				// Worker input buffer full
				p.logger.Warn().
					Str("worker", key).
					Msg("Worker input buffer full, dropping event")
			}
		}
	}

	// Update metrics
	p.metricsMu.Lock()
	p.totalEvents++
	p.metricsMu.Unlock()
}

// collectMetrics aggregates candles from workers and forwards them
func (p *Pool) collectMetrics() {
	defer p.wg.Done()

	p.logger.Info().Msg("Metrics collector started")

	// Collect candles from all workers
	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info().Msg("Metrics collector stopping")
			return

		default:
			p.collectCandlesFromWorkers()
			time.Sleep(10 * time.Millisecond) // REQ-032: Low latency collection
		}
	}
}

// collectCandlesFromWorkers gathers candles from all workers
func (p *Pool) collectCandlesFromWorkers() {
	p.workersMu.RLock()
	workers := make([]*SymbolWorker, 0, len(p.workers))
	for _, worker := range p.workers {
		workers = append(workers, worker)
	}
	p.workersMu.RUnlock()

	for _, worker := range workers {
		select {
		case candle := <-worker.Output:
			// Forward candle to output
			select {
			case p.candleOutput <- candle:
				p.metricsMu.Lock()
				p.totalCandles++
				p.metricsMu.Unlock()
			default:
				p.logger.Warn().
					Str("symbol", candle.Symbol).
					Str("interval", candle.Interval).
					Msg("Candle output buffer full, dropping candle")
			}
		default:
			// No candle available from this worker
		}
	}
}

// healthCheck monitors worker health and performance
func (p *Pool) healthCheck() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	p.logger.Info().Msg("Health checker started")

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info().Msg("Health checker stopping")
			return

		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

// performHealthCheck checks the health of all workers
func (p *Pool) performHealthCheck() {
	p.workersMu.RLock()
	workerCount := len(p.workers)
	p.workersMu.RUnlock()

	p.metricsMu.RLock()
	totalEvents := p.totalEvents
	totalCandles := p.totalCandles
	p.metricsMu.RUnlock()

	p.logger.Info().
		Int("active_workers", workerCount).
		Int64("total_events", totalEvents).
		Int64("total_candles", totalCandles).
		Int("event_queue_size", len(p.eventInput)).
		Int("candle_queue_size", len(p.candleOutput)).
		Msg("Worker pool health check")
}

// GetMetrics returns pool performance metrics
func (p *Pool) GetMetrics() map[string]interface{} {
	p.workersMu.RLock()
	workerCount := len(p.workers)
	workerDetails := make(map[string]interface{})
	for key, worker := range p.workers {
		workerDetails[key] = worker.GetStatus()
	}
	p.workersMu.RUnlock()

	p.metricsMu.RLock()
	totalEvents := p.totalEvents
	totalCandles := p.totalCandles
	p.metricsMu.RUnlock()

	return map[string]interface{}{
		"active_workers":    workerCount,
		"total_events":      totalEvents,
		"total_candles":     totalCandles,
		"event_queue_size":  len(p.eventInput),
		"candle_queue_size": len(p.candleOutput),
		"max_workers":       p.config.MaxWorkers,
		"worker_details":    workerDetails,
	}
}

// GetStatus returns the current status of the pool
func (p *Pool) GetStatus() string {
	select {
	case <-p.ctx.Done():
		return "stopped"
	default:
		return "running"
	}
}
