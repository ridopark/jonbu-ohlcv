package stream

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// REQ-017: WebSocket server setup and client management
// REQ-018: Symbol and timeframe subscription management
// REQ-020: Backpressure and flow control

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Client subscriptions by symbol:timeframe
	subscriptions map[string]map[*Client]bool

	// Inbound messages from clients
	register   chan *Client
	unregister chan *Client
	subscribe  chan SubscriptionEvent

	// Broadcast channel for OHLCV data
	broadcast         chan CandleBroadcast
	enrichedBroadcast chan EnrichedCandleBroadcast

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	mu           sync.RWMutex
	clientCount  int
	messageCount int64
	logger       zerolog.Logger
}

// SubscriptionEvent represents subscription changes
type SubscriptionEvent struct {
	Client    *Client
	Symbol    string
	Timeframe string
	Action    string // subscribe, unsubscribe
}

// CandleBroadcast represents candle data to broadcast
type CandleBroadcast struct {
	Symbol    string
	Timeframe string
	Candle    *models.Candle
}

// EnrichedCandleBroadcast represents enriched candle data to broadcast
type EnrichedCandleBroadcast struct {
	Symbol    string
	Timeframe string
	Candle    *models.EnrichedCandle
}

// NewHub creates a new WebSocket hub
func NewHub(logger zerolog.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:           make(map[*Client]bool),
		subscriptions:     make(map[string]map[*Client]bool),
		register:          make(chan *Client, 100),
		unregister:        make(chan *Client, 100),
		subscribe:         make(chan SubscriptionEvent, 1000),
		broadcast:         make(chan CandleBroadcast, 10000),         // REQ-020: Large buffer for backpressure
		enrichedBroadcast: make(chan EnrichedCandleBroadcast, 10000), // REQ-020: Large buffer for enriched candles
		ctx:               ctx,
		cancel:            cancel,
		logger: logger.With().
			Str("component", "websocket_hub").
			Logger(),
	}
}

// Start begins the hub's main loop
func (h *Hub) Start() {
	h.logger.Info().Msg("WebSocket hub started")

	go h.run()
}

// Stop gracefully shuts down the hub
func (h *Hub) Stop() {
	h.logger.Info().Msg("Stopping WebSocket hub")
	h.cancel()

	// Close all client connections
	h.mu.Lock()
	for client := range h.clients {
		close(client.send)
	}
	h.mu.Unlock()
}

// run is the main hub loop handling client registration and message broadcasting
func (h *Hub) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info().Msg("WebSocket hub shutting down")
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case event := <-h.subscribe:
			h.handleSubscription(event)

		case broadcast := <-h.broadcast:
			h.broadcastCandle(broadcast)

		case enrichedBroadcast := <-h.enrichedBroadcast:
			h.broadcastEnrichedCandle(enrichedBroadcast)

		case <-ticker.C:
			h.logMetrics()
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.clientCount++

	h.logger.Info().
		Str("client_id", client.ID).
		Int("total_clients", h.clientCount).
		Msg("Client registered")

	// Send welcome message
	client.sendMessage(ServerMessage{
		Type:      "connected",
		Timestamp: time.Now(),
	})
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		h.clientCount--

		// Remove client from all subscriptions
		for subscriptionKey, clients := range h.subscriptions {
			if _, subscribed := clients[client]; subscribed {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.subscriptions, subscriptionKey)
				}
			}
		}

		close(client.send)

		h.logger.Info().
			Str("client_id", client.ID).
			Int("total_clients", h.clientCount).
			Msg("Client unregistered")
	}
}

// handleSubscription manages client subscriptions
func (h *Hub) handleSubscription(event SubscriptionEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscriptionKey := event.Symbol + ":" + event.Timeframe

	switch event.Action {
	case "subscribe":
		if h.subscriptions[subscriptionKey] == nil {
			h.subscriptions[subscriptionKey] = make(map[*Client]bool)
		}
		h.subscriptions[subscriptionKey][event.Client] = true

		h.logger.Debug().
			Str("client_id", event.Client.ID).
			Str("subscription", subscriptionKey).
			Int("subscribers", len(h.subscriptions[subscriptionKey])).
			Msg("Client subscribed")

	case "unsubscribe":
		if clients, exists := h.subscriptions[subscriptionKey]; exists {
			delete(clients, event.Client)
			if len(clients) == 0 {
				delete(h.subscriptions, subscriptionKey)
			}

			h.logger.Debug().
				Str("client_id", event.Client.ID).
				Str("subscription", subscriptionKey).
				Msg("Client unsubscribed")
		}
	}
}

// broadcastCandle sends candle data to subscribed clients
func (h *Hub) broadcastCandle(broadcast CandleBroadcast) {
	h.mu.RLock()
	subscriptionKey := broadcast.Symbol + ":" + broadcast.Timeframe
	clients, exists := h.subscriptions[subscriptionKey]
	h.mu.RUnlock()

	if !exists || len(clients) == 0 {
		// No subscribers for this symbol:timeframe
		return
	}

	h.messageCount++

	message := ServerMessage{
		Type:      "candle",
		Symbol:    broadcast.Symbol,
		Timeframe: broadcast.Timeframe,
		Interval:  broadcast.Timeframe, // Add interval at top level
		Data:      broadcast.Candle,
		Timestamp: time.Now(),
	}

	// Broadcast to all subscribed clients
	sentCount := 0
	for client := range clients {
		select {
		case <-h.ctx.Done():
			return
		default:
			client.sendMessage(message)
			sentCount++
		}
	}

	h.logger.Debug().
		Str("symbol", broadcast.Symbol).
		Str("timeframe", broadcast.Timeframe).
		Int("clients", sentCount).
		Msg("Broadcasted candle data")
}

// broadcastEnrichedCandle sends enriched candle data to subscribed clients
func (h *Hub) broadcastEnrichedCandle(broadcast EnrichedCandleBroadcast) {
	// Debug: Log what we're broadcasting
	h.logger.Info().
		Str("symbol", broadcast.Symbol).
		Str("timeframe", broadcast.Timeframe).
		Msg("Broadcasting enriched candle - DEBUG")

	subscriptionKey := fmt.Sprintf("%s:%s", broadcast.Symbol, broadcast.Timeframe)

	h.mu.RLock()
	clients, exists := h.subscriptions[subscriptionKey]
	if !exists || len(clients) == 0 {
		h.mu.RUnlock()
		return
	}
	h.mu.RUnlock()

	// Create WebSocket message for enriched candle with interval at top level
	enrichedData := map[string]interface{}{
		"ohlcv":      broadcast.Candle.OHLCV,
		"indicators": broadcast.Candle.Indicators,
		"analysis":   broadcast.Candle.Analysis,
		"signals":    broadcast.Candle.Signals,
		"metadata":   broadcast.Candle.Metadata,
		"interval":   broadcast.Timeframe, // Add interval at top level for frontend
	}

	message := ServerMessage{
		Type:      "enriched_candle",
		Symbol:    broadcast.Symbol,
		Timeframe: broadcast.Timeframe,
		Interval:  broadcast.Timeframe, // Add interval at top level
		Data:      enrichedData,
		Timestamp: time.Now(),
	}

	// Debug: Log the actual message being sent
	h.logger.Info().
		Str("message_type", message.Type).
		Str("message_symbol", message.Symbol).
		Str("message_timeframe", message.Timeframe).
		Str("message_interval", message.Interval).
		Interface("enriched_data_interval", enrichedData["interval"]).
		Msg("About to send enriched candle message - DEBUG")

	// Broadcast to all subscribed clients
	sentCount := 0
	for client := range clients {
		select {
		case <-h.ctx.Done():
			return
		default:
			client.sendMessage(message)
			sentCount++
		}
	}

	h.logger.Debug().
		Str("symbol", broadcast.Symbol).
		Str("timeframe", broadcast.Timeframe).
		Int("clients", sentCount).
		Msg("Broadcasted enriched candle data")
}

// BroadcastCandle queues a candle for broadcasting
func (h *Hub) BroadcastCandle(symbol, timeframe string, candle *models.Candle) {
	select {
	case h.broadcast <- CandleBroadcast{
		Symbol:    symbol,
		Timeframe: timeframe,
		Candle:    candle,
	}:
	default:
		// REQ-020: Drop messages if broadcast buffer is full
		h.logger.Warn().
			Str("symbol", symbol).
			Str("timeframe", timeframe).
			Msg("Broadcast buffer full, dropping candle")
	}
}

// BroadcastEnrichedCandle queues an enriched candle for broadcasting
func (h *Hub) BroadcastEnrichedCandle(symbol, timeframe string, candle *models.EnrichedCandle) {
	select {
	case h.enrichedBroadcast <- EnrichedCandleBroadcast{
		Symbol:    symbol,
		Timeframe: timeframe,
		Candle:    candle,
	}:
	default:
		// REQ-020: Drop messages if broadcast buffer is full
		h.logger.Warn().
			Str("symbol", symbol).
			Str("timeframe", timeframe).
			Msg("Enriched broadcast buffer full, dropping enriched candle")
	}
}

// RegisterClient adds a client to the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient removes a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// GetMetrics returns hub metrics
func (h *Hub) GetMetrics() (clientCount int, messageCount int64, subscriptionCount int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.clientCount, h.messageCount, len(h.subscriptions)
}

// logMetrics periodically logs hub metrics
func (h *Hub) logMetrics() {
	clientCount, messageCount, subscriptionCount := h.GetMetrics()

	h.logger.Info().
		Int("clients", clientCount).
		Int64("messages_sent", messageCount).
		Int("active_subscriptions", subscriptionCount).
		Msg("Hub metrics")
}
