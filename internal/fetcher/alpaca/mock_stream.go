package alpaca

import (
	"context"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// MockStreamClient simulates Alpaca streaming for testing and development
// Generates 1-minute OHLCV candles at accelerated speeds for testing
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

	// Mock configuration for OHLCV generation
	candleInterval  time.Duration // How often to generate new candles (e.g., 5s instead of 60s)
	priceVariance   float64       // Price volatility for realistic movements
	speedMultiplier float64       // Speed multiplier for testing

	logger zerolog.Logger
}

// MockSymbolData holds mock OHLCV state for a symbol
type MockSymbolData struct {
	Symbol       string
	BasePrice    float64
	CurrentPrice float64
	LastCandle   models.Candle
	LastUpdate   time.Time
	Trend        float64 // -1 to 1, affects price movement direction
	Volume       int64   // Accumulated volume for current period
}

// NewMockStreamClient creates a new mock Alpaca streaming client
func NewMockStreamClient(apiKey, secretKey, baseURL string, logger zerolog.Logger) *MockStreamClient {
	ctx, cancel := context.WithCancel(context.Background())

	// Read configuration from environment variables
	speedMultiplier := 10.0 // Default 10x speed
	if envSpeed := os.Getenv("ALPACA_MOCK_SPEED_MULTIPLIER"); envSpeed != "" {
		if parsed, err := strconv.ParseFloat(envSpeed, 64); err == nil && parsed > 0 {
			speedMultiplier = parsed
		}
	}

	candleIntervalSec := 6 // Default 6 seconds (10x speed for 1-minute candles)
	if envInterval := os.Getenv("ALPACA_MOCK_CANDLE_INTERVAL_SEC"); envInterval != "" {
		if parsed, err := strconv.Atoi(envInterval); err == nil && parsed > 0 {
			candleIntervalSec = parsed
		}
	}

	candleInterval := time.Duration(candleIntervalSec) * time.Second

	logger.Info().
		Float64("speed_multiplier", speedMultiplier).
		Dur("candle_interval", candleInterval).
		Msg("Mock client configuration loaded")

	return &MockStreamClient{
		apiKey:          apiKey,
		secretKey:       secretKey,
		baseURL:         baseURL,
		output:          make(chan models.MarketEvent, 10000),
		symbols:         make(map[string]*MockSymbolData),
		ctx:             ctx,
		cancel:          cancel,
		candleInterval:  candleInterval,
		priceVariance:   0.02, // 2% price variance
		speedMultiplier: speedMultiplier,
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
				LastCandle: models.Candle{
					Symbol:     symbol,
					Timestamp:  time.Now().Truncate(time.Minute),
					Open:       basePrice,
					High:       basePrice,
					Low:        basePrice,
					Close:      basePrice,
					Volume:     0,
					Interval:   "1m",
					LastUpdate: time.Now(),
				},
				LastUpdate: time.Now(),
				Trend:      (rand.Float64() - 0.5) * 2, // Random initial trend
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

// SetSpeed adjusts the mock OHLCV candle generation speed
func (m *MockStreamClient) SetSpeed(multiplier float64) {
	if multiplier <= 0 {
		multiplier = 1.0
	}

	m.speedMultiplier = multiplier
	baseInterval := 60 * time.Second // Base 1-minute interval
	m.candleInterval = time.Duration(float64(baseInterval) / multiplier)

	m.logger.Info().
		Float64("multiplier", multiplier).
		Dur("new_interval", m.candleInterval).
		Msg("Mock candle generation speed updated")
}

// SetCandleInterval directly sets the candle generation interval
func (m *MockStreamClient) SetCandleInterval(interval time.Duration) {
	if interval < time.Second {
		interval = time.Second // Minimum 1 second
	}

	m.candleInterval = interval
	m.speedMultiplier = float64(60*time.Second) / float64(interval)

	m.logger.Info().
		Dur("interval", interval).
		Float64("speed_multiplier", m.speedMultiplier).
		Msg("Mock candle interval updated")
}

// GetSpeed returns the current speed multiplier
func (m *MockStreamClient) GetSpeed() float64 {
	return m.speedMultiplier
}

// GetCandleInterval returns the current candle generation interval
func (m *MockStreamClient) GetCandleInterval() time.Duration {
	return m.candleInterval
}

// generateMockData creates realistic OHLCV candles at accelerated intervals
func (m *MockStreamClient) generateMockData() {
	ticker := time.NewTicker(m.candleInterval)
	defer ticker.Stop()

	m.logger.Info().
		Dur("interval", m.candleInterval).
		Float64("speed_multiplier", m.speedMultiplier).
		Msg("Starting mock OHLCV candle generation")

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if !m.running {
				return
			}

			// Generate OHLCV candles for each subscribed symbol
			for _, symbolData := range m.symbols {
				candle := m.generateMockCandle(symbolData)

				// Convert OHLCV candle to MarketEvent (bar type)
				event := models.MarketEvent{
					Type:      "bar", // OHLCV bar type
					Symbol:    candle.Symbol,
					Timestamp: candle.Timestamp,
					Price:     candle.Close, // Use close price as primary price
					Volume:    candle.Volume,
				}

				select {
				case m.output <- event:
					m.logger.Debug().
						Str("symbol", symbolData.Symbol).
						Float64("open", candle.Open).
						Float64("high", candle.High).
						Float64("low", candle.Low).
						Float64("close", candle.Close).
						Int64("volume", candle.Volume).
						Time("timestamp", candle.Timestamp).
						Msg("Generated mock OHLCV candle")
				case <-m.ctx.Done():
					return
				default:
					// Channel full, drop the event and log warning
					m.logger.Warn().
						Str("symbol", symbolData.Symbol).
						Msg("Event channel full, dropping OHLCV candle")
				}
			}
		}
	}
}

// generateMockCandle creates a realistic OHLCV candle for a symbol
func (m *MockStreamClient) generateMockCandle(data *MockSymbolData) models.Candle {
	now := time.Now()

	// Simulate market volatility with random walk
	change := (rand.Float64() - 0.5) * 0.02 // 2% max change per candle

	// Apply trend bias
	change += data.Trend * 0.005 // Trend influences direction

	// Calculate OHLCV based on current price and change
	open := data.CurrentPrice
	high := open
	low := open
	close := open * (1 + change)

	// Ensure close is within reasonable bounds
	if close <= 0 {
		close = open * 0.99 // Minimum 1% down
	}

	// Generate realistic intraday high/low
	volatility := 0.005 + rand.Float64()*0.01 // 0.5% to 1.5% intraday range
	highChange := rand.Float64() * volatility
	lowChange := rand.Float64() * volatility

	high = math.Max(open, close) * (1 + highChange)
	low = math.Min(open, close) * (1 - lowChange)

	// Ensure OHLC consistency
	if high < math.Max(open, close) {
		high = math.Max(open, close)
	}
	if low > math.Min(open, close) {
		low = math.Min(open, close)
	}

	// Generate realistic volume (higher during market hours)
	baseVolume := int64(1000 + rand.Intn(5000))
	if m.isMarketHours(now) {
		baseVolume *= 3 // 3x volume during market hours
	}

	// Add some volume randomness
	volumeMultiplier := 0.5 + rand.Float64()*1.5 // 0.5x to 2x multiplier
	volume := int64(float64(baseVolume) * volumeMultiplier)

	// Update symbol data for next candle
	data.CurrentPrice = close
	data.LastUpdate = now
	data.Volume = volume

	// Randomly adjust trend every few candles
	if rand.Float64() < 0.1 { // 10% chance
		data.Trend = (rand.Float64() - 0.5) * 2 // New trend between -1 and 1
	}

	candle := models.Candle{
		Symbol:     data.Symbol,
		Timestamp:  now.Truncate(time.Minute), // Align to minute boundary
		Open:       open,
		High:       high,
		Low:        low,
		Close:      close,
		Volume:     volume,
		Interval:   "1m",
		LastUpdate: now,
	}

	data.LastCandle = candle
	return candle
}

// generateSymbolEvent creates a realistic market event for a symbol
func (m *MockStreamClient) generateSymbolEvent(data *MockSymbolData) {
	now := time.Now()

	// Simulate market hours (9:30 AM - 4:00 PM EST)
	// For testing, we'll generate data 24/7 but with different volumes
	isMarketHours := m.isMarketHours(now)

	// Generate different types of events with realistic frequencies
	eventType := m.selectEventType()

	switch eventType {
	case "quote":
		m.generateQuoteEvent(data, now, isMarketHours)
	case "trade":
		m.generateTradeEvent(data, now, isMarketHours)
	case "bar":
		// Bar events only on minute boundaries
		if now.Second() == 0 {
			m.generateBarEvent(data, now)
		}
	}

	// Occasionally update trend
	if rand.Float64() < 0.01 { // 1% chance per event
		data.Trend = (rand.Float64() - 0.5) * 2
	}
}

// selectEventType chooses between quote/trade events realistically
func (m *MockStreamClient) selectEventType() string {
	// Real markets: ~80% quotes, ~20% trades
	if rand.Float64() < 0.8 {
		return "quote"
	}
	return "trade"
}

// generateQuoteEvent creates bid/ask quote updates (most common)
func (m *MockStreamClient) generateQuoteEvent(data *MockSymbolData, timestamp time.Time, isMarketHours bool) {
	// For simplicity, we'll generate quote events as trade events with zero volume
	// In a full implementation, you'd extend MarketEvent to include Bid/Ask
	// Generate small price movements for quotes (don't change the actual price)
	spread := 0.01 + rand.Float64()*0.04

	// Quote events show price discovery but don't execute
	quoteEvent := models.MarketEvent{
		Symbol:    data.Symbol,
		Price:     data.CurrentPrice + (rand.Float64()-0.5)*spread,
		Volume:    0, // Quotes don't have volume
		Timestamp: timestamp,
		Type:      "quote",
	}

	// Send quote event
	select {
	case m.output <- quoteEvent:
		// Quote sent successfully
	default:
		// Channel full, drop event
		m.logger.Warn().
			Str("symbol", data.Symbol).
			Msg("Mock output buffer full, dropping quote")
	}
}

// generateTradeEvent creates actual trade events (less common but more impactful)
func (m *MockStreamClient) generateTradeEvent(data *MockSymbolData, timestamp time.Time, isMarketHours bool) {
	// Generate price movement based on trend and random walk
	priceChange := m.generatePriceChange(data)
	newPrice := data.CurrentPrice + priceChange

	// Ensure price doesn't go negative
	if newPrice <= 0 {
		newPrice = data.CurrentPrice * 0.99
	}

	// Generate realistic trade sizes (most trades are small)
	volume := m.generateRealisticTradeSize(data, isMarketHours, math.Abs(priceChange))

	// Update symbol data ONLY on trades
	data.CurrentPrice = newPrice
	data.Volume += volume
	data.LastUpdate = timestamp

	// Create trade event
	tradeEvent := models.MarketEvent{
		Symbol:    data.Symbol,
		Price:     newPrice,
		Volume:    volume,
		Timestamp: timestamp,
		Type:      "trade",
	}

	// Send event (non-blocking)
	select {
	case m.output <- tradeEvent:
		// Trade sent successfully
	default:
		// Channel full, drop event
		m.logger.Warn().
			Str("symbol", data.Symbol).
			Msg("Mock output buffer full, dropping trade")
	}
}

// generateRealisticTradeSize creates realistic trade volumes
func (m *MockStreamClient) generateRealisticTradeSize(data *MockSymbolData, isMarketHours bool, priceChange float64) int64 {
	// Most real trades are very small (1-100 shares)
	// Occasionally large institutional trades (1000+ shares)

	if rand.Float64() < 0.05 { // 5% chance of large trade
		return int64(1000 + rand.Float64()*9000) // 1000-10000 shares
	}

	if rand.Float64() < 0.2 { // 20% chance of medium trade
		return int64(100 + rand.Float64()*900) // 100-1000 shares
	}

	// 75% small trades (1-100 shares) - this is realistic!
	return int64(1 + rand.Float64()*99)
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
