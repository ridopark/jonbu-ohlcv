package alpaca

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// MockStreamClient simulates Alpaca streaming for testing and development
type MockStreamClient struct {
	// Configuration
	apiKey    string
	secretKey string
	baseURL   string

	// Mock state
	connected bool
	running   bool

	// Channels
	output chan models.MarketEvent

	// Subscriptions
	symbols map[string]*MockSymbolData

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Mock configuration
	eventInterval time.Duration
	priceVariance float64

	logger zerolog.Logger
}

// MockSymbolData holds mock data for a symbol
type MockSymbolData struct {
	Symbol       string
	BasePrice    float64
	CurrentPrice float64
	Volume       int64
	LastUpdate   time.Time
	Trend        float64 // -1 to 1, affects price movement direction
}

// NewMockStreamClient creates a new mock Alpaca streaming client
func NewMockStreamClient(apiKey, secretKey, baseURL string, logger zerolog.Logger) *MockStreamClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &MockStreamClient{
		apiKey:        apiKey,
		secretKey:     secretKey,
		baseURL:       baseURL,
		output:        make(chan models.MarketEvent, 10000),
		symbols:       make(map[string]*MockSymbolData),
		ctx:           ctx,
		cancel:        cancel,
		eventInterval: 100 * time.Millisecond, // Generate events every 100ms
		priceVariance: 0.02,                   // 2% price variance
		logger: logger.With().
			Str("component", "mock_alpaca_stream").
			Logger(),
	}
}

// Start begins the mock streaming
func (m *MockStreamClient) Start() error {
	m.logger.Info().Msg("Starting mock Alpaca stream client")

	m.connected = true
	m.running = true

	// Start the mock data generator
	go m.generateMockData()

	m.logger.Info().Msg("Mock Alpaca stream started successfully")
	return nil
}

// Stop gracefully shuts down the mock streaming client
func (m *MockStreamClient) Stop() {
	m.logger.Info().Msg("Stopping mock Alpaca stream client")

	m.running = false
	m.connected = false
	m.cancel()

	close(m.output)
}

// Subscribe adds symbols to the mock stream subscription
func (m *MockStreamClient) Subscribe(symbols []string) error {
	if !m.connected {
		return nil // Mock always succeeds
	}

	for _, symbol := range symbols {
		if _, exists := m.symbols[symbol]; !exists {
			basePrice := m.getBasePrice(symbol)
			m.symbols[symbol] = &MockSymbolData{
				Symbol:       symbol,
				BasePrice:    basePrice,
				CurrentPrice: basePrice,
				Volume:       0,
				LastUpdate:   time.Now(),
				Trend:        (rand.Float64() - 0.5) * 2, // Random initial trend
			}
		}
	}

	m.logger.Info().
		Strs("symbols", symbols).
		Msg("Mock subscribed to symbols")

	return nil
}

// Unsubscribe removes symbols from the mock stream subscription
func (m *MockStreamClient) Unsubscribe(symbols []string) error {
	if !m.connected {
		return nil // Mock always succeeds
	}

	for _, symbol := range symbols {
		delete(m.symbols, symbol)
	}

	m.logger.Info().
		Strs("symbols", symbols).
		Msg("Mock unsubscribed from symbols")

	return nil
}

// GetOutput returns the channel for consuming mock market events
func (m *MockStreamClient) GetOutput() <-chan models.MarketEvent {
	return m.output
}

// GetConnectionStatus returns the mock connection status
func (m *MockStreamClient) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected":          m.connected,
		"subscribed_symbols": len(m.symbols),
		"reconnect_attempts": 0,
		"max_reconnects":     0,
		"mock":               true,
	}
}

// generateMockData creates realistic market data events
func (m *MockStreamClient) generateMockData() {
	ticker := time.NewTicker(m.eventInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if !m.running {
				return
			}

			// Generate events for each subscribed symbol
			for _, symbolData := range m.symbols {
				m.generateSymbolEvent(symbolData)
			}
		}
	}
}

// generateSymbolEvent creates a realistic market event for a symbol
func (m *MockStreamClient) generateSymbolEvent(data *MockSymbolData) {
	now := time.Now()

	// Simulate market hours (9:30 AM - 4:00 PM EST)
	// For testing, we'll generate data 24/7 but with different volumes
	isMarketHours := m.isMarketHours(now)

	// Generate price movement based on trend and random walk
	priceChange := m.generatePriceChange(data)
	newPrice := data.CurrentPrice + priceChange

	// Ensure price doesn't go negative
	if newPrice <= 0 {
		newPrice = data.CurrentPrice * 0.99
	}

	// Generate volume based on market hours and price movement
	volume := m.generateVolume(data, isMarketHours, math.Abs(priceChange))

	// Update symbol data
	data.CurrentPrice = newPrice
	data.Volume += volume
	data.LastUpdate = now

	// Create trade event
	tradeEvent := models.MarketEvent{
		Symbol:    data.Symbol,
		Price:     newPrice,
		Volume:    volume,
		Timestamp: now,
		Type:      "trade",
	}

	// Send event (non-blocking)
	select {
	case m.output <- tradeEvent:
		// Event sent successfully
	default:
		// Channel full, drop event
		m.logger.Warn().
			Str("symbol", data.Symbol).
			Msg("Mock output buffer full, dropping event")
	}

	// Generate periodic bar events (every minute on the minute)
	if now.Second() == 0 && now.Sub(data.LastUpdate) >= time.Minute {
		m.generateBarEvent(data, now)
	}

	// Occasionally update trend
	if rand.Float64() < 0.01 { // 1% chance per event
		data.Trend = (rand.Float64() - 0.5) * 2
	}
}

// generatePriceChange creates realistic price movements
func (m *MockStreamClient) generatePriceChange(data *MockSymbolData) float64 {
	// Random walk with trend bias
	randomComponent := (rand.Float64() - 0.5) * 2 * m.priceVariance * data.BasePrice
	trendComponent := data.Trend * 0.001 * data.BasePrice

	// Mean reversion component (pulls price back to base)
	meanReversionStrength := 0.01
	meanReversionComponent := (data.BasePrice - data.CurrentPrice) * meanReversionStrength

	return randomComponent + trendComponent + meanReversionComponent
}

// generateVolume creates realistic volume based on conditions
func (m *MockStreamClient) generateVolume(data *MockSymbolData, isMarketHours bool, priceChange float64) int64 {
	baseVolume := int64(100)

	// Higher volume during market hours
	if isMarketHours {
		baseVolume *= 5
	}

	// Higher volume with larger price changes
	volatilityMultiplier := 1 + (priceChange/data.BasePrice)*10
	if volatilityMultiplier < 0.1 {
		volatilityMultiplier = 0.1
	}

	// Add randomness
	randomMultiplier := 0.5 + rand.Float64()

	return int64(float64(baseVolume) * volatilityMultiplier * randomMultiplier)
}

// generateBarEvent creates a 1-minute OHLCV bar event
func (m *MockStreamClient) generateBarEvent(data *MockSymbolData, timestamp time.Time) {
	// For simplicity, use current price as OHLC
	// In reality, this would track the actual OHLC over the minute
	barEvent := models.MarketEvent{
		Symbol:    data.Symbol,
		Price:     data.CurrentPrice,
		Volume:    data.Volume,
		Timestamp: timestamp.Truncate(time.Minute),
		Type:      "bar",
	}

	select {
	case m.output <- barEvent:
		// Bar event sent successfully
		m.logger.Debug().
			Str("symbol", data.Symbol).
			Float64("price", data.CurrentPrice).
			Int64("volume", data.Volume).
			Msg("Generated mock bar event")
	default:
		// Channel full, drop event
		m.logger.Warn().
			Str("symbol", data.Symbol).
			Msg("Mock output buffer full, dropping bar event")
	}

	// Reset volume counter for next minute
	data.Volume = 0
}

// isMarketHours checks if it's during market hours
func (m *MockStreamClient) isMarketHours(t time.Time) bool {
	// Convert to EST
	est, _ := time.LoadLocation("America/New_York")
	estTime := t.In(est)

	// Check if it's a weekday
	weekday := estTime.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// Check if it's between 9:30 AM and 4:00 PM EST
	hour := estTime.Hour()
	minute := estTime.Minute()

	// Market opens at 9:30 AM
	if hour < 9 || (hour == 9 && minute < 30) {
		return false
	}

	// Market closes at 4:00 PM
	if hour >= 16 {
		return false
	}

	return true
}

// getBasePrice returns a realistic base price for a symbol
func (m *MockStreamClient) getBasePrice(symbol string) float64 {
	// Set realistic base prices for common symbols
	basePrices := map[string]float64{
		"AAPL":  180.00,
		"GOOGL": 140.00,
		"MSFT":  380.00,
		"TSLA":  250.00,
		"AMZN":  150.00,
		"NVDA":  800.00,
		"META":  350.00,
		"NFLX":  450.00,
		"AMD":   140.00,
		"INTC":  25.00,
	}

	if price, exists := basePrices[symbol]; exists {
		return price
	}

	// Default price for unknown symbols
	return 100.00 + rand.Float64()*400.00 // Random price between $100-$500
}
