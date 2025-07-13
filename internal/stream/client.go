package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// REQ-017: WebSocket endpoints for real-time data streaming
// REQ-018: Symbol and timeframe subscription management
// REQ-019: Connection lifecycle handling
// REQ-020: Backpressure and flow control

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
}

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	conn          *websocket.Conn
	hub           *Hub
	send          chan []byte
	subscriptions map[string]bool // symbol:timeframe keys
	logger        zerolog.Logger
	mu            sync.RWMutex
}

// ClientMessage represents messages from clients
type ClientMessage struct {
	Type      string `json:"type"`
	Symbol    string `json:"symbol,omitempty"`
	Timeframe string `json:"timeframe,omitempty"`
	Action    string `json:"action,omitempty"` // subscribe, unsubscribe
}

// ServerMessage represents messages to clients
type ServerMessage struct {
	Type      string         `json:"type"`
	Symbol    string         `json:"symbol,omitempty"`
	Timeframe string         `json:"timeframe,omitempty"`
	Data      *models.Candle `json:"data,omitempty"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, logger zerolog.Logger) *Client {
	return &Client{
		ID:            generateClientID(),
		conn:          conn,
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		logger: logger.With().
			Str("component", "websocket_client").
			Str("client_id", generateClientID()).
			Logger(),
	}
}

// Start begins the client's read and write goroutines
func (c *Client) Start(ctx context.Context) {
	c.logger.Info().Msg("Client connection started")

	go c.writePump(ctx)
	go c.readPump(ctx)
}

// readPump handles incoming messages from the client
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.logger.Info().Msg("Client read pump closed")
	}()

	// Set read deadline and pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg ClientMessage
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error().Err(err).Msg("WebSocket read error")
				}
				return
			}

			c.handleMessage(msg)
		}
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.logger.Info().Msg("Client write pump closed")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.logger.Error().Err(err).Msg("Failed to get WebSocket writer")
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.logger.Error().Err(err).Msg("Failed to close WebSocket writer")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Error().Err(err).Msg("Failed to send ping")
				return
			}
		}
	}
}

// handleMessage processes incoming client messages
func (c *Client) handleMessage(msg ClientMessage) {
	c.logger.Debug().
		Str("type", msg.Type).
		Str("symbol", msg.Symbol).
		Str("timeframe", msg.Timeframe).
		Str("action", msg.Action).
		Msg("Handling client message")

	switch msg.Type {
	case "subscription":
		c.handleSubscription(msg)
	case "ping":
		c.sendMessage(ServerMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		})
	default:
		c.sendError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handleSubscription manages symbol/timeframe subscriptions
func (c *Client) handleSubscription(msg ClientMessage) {
	if msg.Symbol == "" || msg.Timeframe == "" {
		c.sendError("Symbol and timeframe are required for subscriptions")
		return
	}

	// REQ-025: Input validation
	if err := c.validateSubscription(msg.Symbol, msg.Timeframe); err != nil {
		c.sendError(fmt.Sprintf("Invalid subscription: %v", err))
		return
	}

	subscriptionKey := fmt.Sprintf("%s:%s", msg.Symbol, msg.Timeframe)

	c.mu.Lock()
	defer c.mu.Unlock()

	switch msg.Action {
	case "subscribe":
		c.subscriptions[subscriptionKey] = true
		c.hub.subscribe <- SubscriptionEvent{
			Client:    c,
			Symbol:    msg.Symbol,
			Timeframe: msg.Timeframe,
			Action:    "subscribe",
		}
		c.logger.Info().
			Str("symbol", msg.Symbol).
			Str("timeframe", msg.Timeframe).
			Msg("Client subscribed")

	case "unsubscribe":
		delete(c.subscriptions, subscriptionKey)
		c.hub.subscribe <- SubscriptionEvent{
			Client:    c,
			Symbol:    msg.Symbol,
			Timeframe: msg.Timeframe,
			Action:    "unsubscribe",
		}
		c.logger.Info().
			Str("symbol", msg.Symbol).
			Str("timeframe", msg.Timeframe).
			Msg("Client unsubscribed")

	default:
		c.sendError(fmt.Sprintf("Unknown subscription action: %s", msg.Action))
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to marshal message")
		return
	}

	select {
	case c.send <- data:
	default:
		// REQ-020: Backpressure handling - close slow clients
		close(c.send)
		c.logger.Warn().Msg("Client send buffer full, closing connection")
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	c.sendMessage(ServerMessage{
		Type:      "error",
		Error:     errMsg,
		Timestamp: time.Now(),
	})
}

// IsSubscribed checks if client is subscribed to symbol:timeframe
func (c *Client) IsSubscribed(symbol, timeframe string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subscriptionKey := fmt.Sprintf("%s:%s", symbol, timeframe)
	return c.subscriptions[subscriptionKey]
}

// GetSubscriptions returns all client subscriptions
func (c *Client) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subscriptions := make([]string, 0, len(c.subscriptions))
	for key := range c.subscriptions {
		subscriptions = append(subscriptions, key)
	}
	return subscriptions
}

// validateSubscription validates symbol and timeframe for subscriptions
func (c *Client) validateSubscription(symbol, timeframe string) error {
	// Basic symbol validation (1-5 uppercase letters)
	if len(symbol) == 0 || len(symbol) > 5 {
		return fmt.Errorf("symbol must be 1-5 characters")
	}

	for _, char := range symbol {
		if char < 'A' || char > 'Z' {
			return fmt.Errorf("symbol must contain only uppercase letters")
		}
	}

	// Timeframe validation
	validTimeframes := map[string]bool{
		"1min":  true,
		"5min":  true,
		"15min": true,
		"30min": true,
		"1hour": true,
		"1day":  true,
	}

	if !validTimeframes[timeframe] {
		return fmt.Errorf("invalid timeframe: %s", timeframe)
	}

	return nil
}

// generateClientID creates a unique client identifier
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
