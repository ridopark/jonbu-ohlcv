package alpaca

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// REQ-006: Real-time data streaming from Alpaca
// REQ-007: WebSocket connection management
// REQ-008: Subscription management for symbols
// REQ-009: Error handling and reconnection logic
// REQ-010: Rate limiting and backpressure handling

// StreamClient handles real-time data streaming from Alpaca
type StreamClient struct {
	// Configuration
	apiKey    string
	secretKey string
	baseURL   string

	// WebSocket connection
	conn      *websocket.Conn
	connected bool

	// Channels
	output chan models.MarketEvent

	// Subscriptions
	symbols map[string]bool

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Reconnection
	reconnectAttempts int
	maxReconnects     int
	reconnectDelay    time.Duration

	logger zerolog.Logger
}

// AlpacaStreamMessage represents incoming Alpaca stream messages
type AlpacaStreamMessage struct {
	Type      string          `json:"T"`
	Symbol    string          `json:"S,omitempty"`
	Price     float64         `json:"p,omitempty"`
	Size      int64           `json:"s,omitempty"`
	Timestamp int64           `json:"t,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// AlpacaAuthMessage represents authentication message
type AlpacaAuthMessage struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// AlpacaSubscribeMessage represents subscription message
type AlpacaSubscribeMessage struct {
	Action string   `json:"action"`
	Trades []string `json:"trades,omitempty"`
	Quotes []string `json:"quotes,omitempty"`
	Bars   []string `json:"bars,omitempty"`
}

// NewStreamClient creates a new Alpaca streaming client
func NewStreamClient(apiKey, secretKey, baseURL string, logger zerolog.Logger) *StreamClient {
	ctx, cancel := context.WithCancel(context.Background())

	// Use proper WebSocket URL for paper trading with IEX data
	streamURL := "wss://stream.data.alpaca.markets/v2/iex"
	if !strings.Contains(baseURL, "paper") {
		// Use SIP data for live trading
		streamURL = "wss://stream.data.alpaca.markets/v2/sip"
	}

	return &StreamClient{
		apiKey:         apiKey,
		secretKey:      secretKey,
		baseURL:        streamURL,
		output:         make(chan models.MarketEvent, 10000), // REQ-010: Large buffer
		symbols:        make(map[string]bool),
		ctx:            ctx,
		cancel:         cancel,
		maxReconnects:  10,
		reconnectDelay: 5 * time.Second,
		logger: logger.With().
			Str("component", "alpaca_stream").
			Logger(),
	}
}

// Start begins the streaming connection
func (c *StreamClient) Start() error {
	c.logger.Info().Msg("Starting Alpaca stream client")

	if err := c.connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	go c.readLoop()
	go c.pingLoop()

	return nil
}

// Stop gracefully shuts down the streaming client
func (c *StreamClient) Stop() {
	c.logger.Info().Msg("Stopping Alpaca stream client")

	c.cancel()

	if c.conn != nil {
		c.conn.Close()
	}

	close(c.output)
}

// Subscribe adds symbols to the stream subscription
func (c *StreamClient) Subscribe(symbols []string) error {
	if !c.connected {
		return fmt.Errorf("not connected to stream")
	}

	// Add symbols to subscription map
	for _, symbol := range symbols {
		c.symbols[symbol] = true
	}

	// Send subscription message
	subscribeMsg := AlpacaSubscribeMessage{
		Action: "subscribe",
		Trades: symbols, // Subscribe to trades for real-time price data
		Bars:   symbols, // Subscribe to 1-minute bars
	}

	if err := c.conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("failed to send subscription: %w", err)
	}

	c.logger.Info().
		Strs("symbols", symbols).
		Msg("Subscribed to symbols")

	return nil
}

// Unsubscribe removes symbols from the stream subscription
func (c *StreamClient) Unsubscribe(symbols []string) error {
	if !c.connected {
		return fmt.Errorf("not connected to stream")
	}

	// Remove symbols from subscription map
	for _, symbol := range symbols {
		delete(c.symbols, symbol)
	}

	// Send unsubscription message
	unsubscribeMsg := AlpacaSubscribeMessage{
		Action: "unsubscribe",
		Trades: symbols,
		Bars:   symbols,
	}

	if err := c.conn.WriteJSON(unsubscribeMsg); err != nil {
		return fmt.Errorf("failed to send unsubscription: %w", err)
	}

	c.logger.Info().
		Strs("symbols", symbols).
		Msg("Unsubscribed from symbols")

	return nil
}

// GetOutput returns the channel for consuming market events
func (c *StreamClient) GetOutput() <-chan models.MarketEvent {
	return c.output
}

// connect establishes WebSocket connection and authenticates
func (c *StreamClient) connect() error {
	c.logger.Info().Str("url", c.baseURL).Msg("Connecting to Alpaca stream")

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid stream URL: %w", err)
	}

	// Establish WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	c.conn = conn

	// Authenticate
	authMsg := AlpacaAuthMessage{
		Action: "auth",
		Key:    c.apiKey,
		Secret: c.secretKey,
	}

	if err := c.conn.WriteJSON(authMsg); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	// Wait for auth response
	if err := c.waitForAuth(); err != nil {
		c.conn.Close()
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.connected = true
	c.reconnectAttempts = 0

	c.logger.Info().Msg("Connected and authenticated to Alpaca stream")
	return nil
}

// waitForAuth waits for authentication confirmation
func (c *StreamClient) waitForAuth() error {
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer c.conn.SetReadDeadline(time.Time{})

	var msg json.RawMessage
	if err := c.conn.ReadJSON(&msg); err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	c.logger.Info().
		RawJSON("auth_response", msg).
		Msg("Received authentication response")

	// Try to parse as array first (Alpaca may send an array of messages)
	var responseArray []map[string]interface{}
	if err := json.Unmarshal(msg, &responseArray); err == nil && len(responseArray) > 0 {
		// Check each message in the array
		for _, response := range responseArray {
			if msgType, ok := response["T"].(string); ok {
				if msgType == "success" {
					c.logger.Info().Msg("Authentication successful")
					return nil
				}
				if msgType == "error" {
					if errMsg, ok := response["msg"].(string); ok {
						return fmt.Errorf("authentication failed: %s", errMsg)
					}
					return fmt.Errorf("authentication failed: %v", response)
				}
			}
		}
	}

	// Try to parse as single object
	var response map[string]interface{}
	if err := json.Unmarshal(msg, &response); err == nil {
		if msgType, ok := response["T"].(string); ok {
			if msgType == "success" {
				c.logger.Info().Msg("Authentication successful")
				return nil
			}
			if msgType == "error" {
				if errMsg, ok := response["msg"].(string); ok {
					return fmt.Errorf("authentication failed: %s", errMsg)
				}
				return fmt.Errorf("authentication failed: %v", response)
			}
		}
	}

	return fmt.Errorf("unexpected auth response format: %s", string(msg))
}

// readLoop continuously reads messages from the WebSocket
func (c *StreamClient) readLoop() {
	defer func() {
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
		c.logger.Info().Msg("Stream read loop ended")
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			var rawMsg json.RawMessage
			if err := c.conn.ReadJSON(&rawMsg); err != nil {
				c.logger.Error().Err(err).Msg("Failed to read stream message")

				// Attempt reconnection
				if c.reconnectAttempts < c.maxReconnects {
					c.attemptReconnect()
				}
				return
			}

			// Try to parse as array first
			var msgArray []AlpacaStreamMessage
			if err := json.Unmarshal(rawMsg, &msgArray); err == nil {
				// Process each message in the array
				for _, msg := range msgArray {
					c.processMessage(msg)
				}
				continue
			}

			// Try to parse as single message
			var msg AlpacaStreamMessage
			if err := json.Unmarshal(rawMsg, &msg); err == nil {
				c.processMessage(msg)
				continue
			}

			// Log unparseable messages for debugging
			c.logger.Debug().
				RawJSON("raw_message", rawMsg).
				Msg("Received unparseable stream message")
		}
	}
}

// processMessage converts Alpaca messages to MarketEvent
func (c *StreamClient) processMessage(msg AlpacaStreamMessage) {
	switch msg.Type {
	case "t": // Trade
		if msg.Symbol == "" || msg.Price <= 0 {
			return
		}

		event := models.MarketEvent{
			Symbol:    msg.Symbol,
			Price:     msg.Price,
			Volume:    msg.Size,
			Timestamp: time.Unix(0, msg.Timestamp*int64(time.Millisecond)),
			Type:      "trade",
		}

		select {
		case c.output <- event:
			// Event sent successfully
		default:
			// REQ-010: Handle backpressure
			c.logger.Warn().
				Str("symbol", msg.Symbol).
				Msg("Output buffer full, dropping trade event")
		}

	case "b": // Bar (1-minute OHLCV)
		if msg.Symbol == "" {
			return
		}

		// For bars, we need to parse the data field
		// This is a simplified version - real implementation would parse OHLCV data
		event := models.MarketEvent{
			Symbol:    msg.Symbol,
			Price:     msg.Price,
			Volume:    msg.Size,
			Timestamp: time.Unix(0, msg.Timestamp*int64(time.Millisecond)),
			Type:      "bar",
		}

		select {
		case c.output <- event:
			// Event sent successfully
		default:
			c.logger.Warn().
				Str("symbol", msg.Symbol).
				Msg("Output buffer full, dropping bar event")
		}

	case "error":
		c.logger.Error().
			RawJSON("data", msg.Data).
			Msg("Received error from Alpaca stream")

	default:
		// Ignore other message types (connection, subscription confirmations, etc.)
		c.logger.Debug().
			Str("type", msg.Type).
			Msg("Received stream message")
	}
}

// attemptReconnect tries to reconnect to the stream
func (c *StreamClient) attemptReconnect() {
	c.reconnectAttempts++

	c.logger.Warn().
		Int("attempt", c.reconnectAttempts).
		Int("max_attempts", c.maxReconnects).
		Msg("Attempting to reconnect to Alpaca stream")

	time.Sleep(c.reconnectDelay)

	if err := c.connect(); err != nil {
		c.logger.Error().Err(err).Msg("Reconnection failed")
		return
	}

	// Re-subscribe to all symbols
	if len(c.symbols) > 0 {
		symbols := make([]string, 0, len(c.symbols))
		for symbol := range c.symbols {
			symbols = append(symbols, symbol)
		}

		if err := c.Subscribe(symbols); err != nil {
			c.logger.Error().Err(err).Msg("Failed to re-subscribe after reconnection")
		}
	}

	// Restart read loop
	go c.readLoop()
}

// pingLoop sends periodic ping messages to keep connection alive
func (c *StreamClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if c.connected && c.conn != nil {
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.logger.Error().Err(err).Msg("Failed to send ping")
				}
			}
		}
	}
}

// GetConnectionStatus returns the current connection status
func (c *StreamClient) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected":          c.connected,
		"subscribed_symbols": len(c.symbols),
		"reconnect_attempts": c.reconnectAttempts,
		"max_reconnects":     c.maxReconnects,
	}
}
