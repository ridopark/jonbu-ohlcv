# Copilot Code Generation Guidelines

## How to Use Requirements

When GitHub Copilot generates code for this project, it should:

### 1. Requirement-Driven Development
- Reference specific requirements (REQ-XXX) in code comments
- Ensure all generated code addresses documented requirements
- Validate that solutions meet performance and reliability criteria
- Include appropriate error handling for each requirement category

### 2. Code Templates to Follow

#### Data Ingestion Components
```go
// REQ-001, REQ-003: Real-time ingestion with reconnection
type DataIngester struct {
    logger    zerolog.Logger
    client    WebSocketClient
    reconnect ReconnectionStrategy
}

func (d *DataIngester) Start(ctx context.Context) error {
    d.logger.Info().Msg("Starting data ingestion")
    // Implementation with proper error handling
}
```

#### Aggregation Workers
```go
// REQ-009, REQ-031: Per-symbol workers with high throughput
type AggregatorWorker struct {
    symbol   string
    interval time.Duration
    logger   zerolog.Logger
    input    <-chan MarketEvent
    output   chan<- Candle
}

func (w *AggregatorWorker) Process() error {
    // REQ-032: Sub-millisecond latency requirement
    start := time.Now()
    defer func() {
        w.logger.Debug().
            Dur("processing_time", time.Since(start)).
            Msg("Event processed")
    }()
    // Implementation
}
```

#### API Handlers
```go
// REQ-041, REQ-018: Input validation and subscription support
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
    // REQ-047: Correlation ID for tracing
    correlationID := uuid.New().String()
    logger := h.logger.With().
        Str("correlation_id", correlationID).
        Logger()
    
    // REQ-041: Input validation
    symbol := mux.Vars(r)["symbol"]
    if err := validateSymbol(symbol); err != nil {
        logger.Error().Err(err).Msg("Invalid symbol")
        http.Error(w, "Invalid symbol", http.StatusBadRequest)
        return
    }
    
    // Implementation
}
```

### 3. Mandatory Patterns

#### Error Handling (REQ-039, REQ-048)
```go
func processData(data *MarketData) error {
    if err := validateData(data); err != nil {
        return fmt.Errorf("data validation failed: %w", err)
    }
    
    if err := storeData(data); err != nil {
        log.Error().
            Err(err).
            Str("symbol", data.Symbol).
            Msg("Failed to store market data")
        return fmt.Errorf("storage failed: %w", err)
    }
    
    return nil
}
```

#### Testing Pattern (REQ-051, REQ-054)
```go
func TestAggregatorWorker_Process(t *testing.T) {
    tests := []struct {
        name     string
        input    MarketEvent
        expected Candle
        wantErr  bool
    }{
        {
            name: "valid market event",
            input: MarketEvent{
                Symbol: "AAPL",
                Price:  150.0,
                Volume: 1000,
            },
            wantErr: false,
        },
        // Additional test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Configuration (REQ-062, REQ-065)
```go
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    Alpaca   AlpacaConfig   `mapstructure:"alpaca"`
    Server   ServerConfig   `mapstructure:"server"`
}

func (c *Config) Validate() error {
    if c.Database.Host == "" {
        return errors.New("database host is required")
    }
    // Additional validation
    return nil
}

func (c *Config) String() string {
    // REQ-065: Mask sensitive values
    masked := *c
    masked.Database.Password = "***"
    masked.Alpaca.SecretKey = "***"
    return fmt.Sprintf("%+v", masked)
}
```

### 4. Architecture Compliance

#### Enriched Candle Strategy: On-Demand Calculation (REQ-200-205)
```go
// IMPORTANT: Enriched candles are NOT stored in database
// They are calculated on-demand with <1ms latency for optimal performance
type BacktestingEngine struct {
    enricher *enrichment.CandleEnrichmentEngine
    ohlcvRepo *database.OHLCVRepository
    logger   zerolog.Logger
}

func (b *BacktestingEngine) GetEnrichedCandles(symbol string, from, to time.Time) ([]*models.EnrichedCandle, error) {
    // 1. Fetch raw OHLCV from database (only raw data is stored)
    ohlcvData, err := b.ohlcvRepo.GetHistory(symbol, from, to)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch OHLCV data: %w", err)
    }
    
    // 2. Calculate enriched candles on-demand (0.167ms per candle)
    enriched := make([]*models.EnrichedCandle, 0, len(ohlcvData))
    for i, candle := range ohlcvData {
        history := ohlcvData[:i]  // Historical context for indicators
        
        enrichedCandle, err := b.enricher.EnrichCandle(ctx, candle, history, nil)
        if err != nil {
            b.logger.Warn().Err(err).Str("symbol", symbol).Msg("Failed to enrich candle")
            continue
        }
        
        enriched = append(enriched, enrichedCandle)
    }
    
    return enriched, nil
}

// Database Repository: Only store raw OHLCV data
type OHLCVRepository struct {
    db *sql.DB
}

func (r *OHLCVRepository) Insert(ohlcv *models.OHLCV) error {
    // REQ-013: Store only basic OHLCV data - enrichment calculated on-demand
    const query = `
        INSERT INTO ohlcv (symbol, timestamp, open, high, low, close, volume, timeframe)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (symbol, timestamp, timeframe) DO NOTHING`
    
    _, err := r.db.Exec(query, ohlcv.Symbol, ohlcv.Timestamp, 
        ohlcv.Open, ohlcv.High, ohlcv.Low, ohlcv.Close, ohlcv.Volume, ohlcv.Timeframe)
    return err
}
```

**Rationale for On-Demand Strategy:**
- **Performance**: 0.167ms enrichment latency makes real-time calculation viable
- **Storage Efficiency**: Avoid 50x storage increase from storing enriched data  
- **Flexibility**: Easy to modify indicators without schema migrations
- **Accuracy**: Always use latest enrichment logic for consistent results
- **Cost**: Minimal computational overhead vs massive storage costs

#### Interface Design (REQ-029)
```go
// Data provider abstraction
type MarketDataProvider interface {
    Connect(ctx context.Context) error
    Subscribe(symbols []string) error
    Events() <-chan MarketEvent
    Close() error
}

// Implementation must be swappable
type AlpacaProvider struct {
    // Implementation
}

func (a *AlpacaProvider) Connect(ctx context.Context) error {
    // REQ-003: Connection with retry logic
}
```

#### Dependency Injection (REQ-027)
```go
type Services struct {
    Logger     zerolog.Logger
    DB         *sql.DB
    Provider   MarketDataProvider
    Aggregator *AggregatorManager
}

func NewServices(cfg *Config) (*Services, error) {
    // Wire dependencies
    return &Services{
        Logger:     initLogger(cfg.LogLevel),
        DB:         initDB(cfg.Database),
        Provider:   NewAlpacaProvider(cfg.Alpaca),
        Aggregator: NewAggregatorManager(),
    }, nil
}
```

### 5. Performance Guidelines

#### Channel Usage (REQ-034)
```go
// Use buffered channels to prevent blocking
events := make(chan MarketEvent, 1000)  // REQ-034
candles := make(chan Candle, 500)

// Monitor channel capacity
func monitorChannelHealth(ch chan MarketEvent) {
    go func() {
        for {
            usage := float64(len(ch)) / float64(cap(ch))
            if usage > 0.8 {
                log.Warn().
                    Float64("usage", usage).
                    Msg("Channel approaching capacity")
            }
            time.Sleep(time.Second)
        }
    }()
}
```

#### Memory Management (REQ-035)
```go
func (w *Worker) processWithCleanup() {
    defer func() {
        // Clean up resources
        w.buffer = w.buffer[:0]  // Reset slice but keep capacity
        runtime.GC()             // Suggest garbage collection if needed
    }()
    
    // Processing logic
}
```

### 6. Monitoring Integration (REQ-081-084)

```go
type HealthChecker struct {
    db       *sql.DB
    provider MarketDataProvider
}

func (h *HealthChecker) Check() map[string]bool {
    return map[string]bool{
        "database":   h.checkDatabase(),
        "provider":   h.checkProvider(),
        "memory":     h.checkMemoryUsage(),
        "goroutines": h.checkGoroutineCount(),
    }
}
```

## Code Generation Checklist

For every generated code block, verify:
- [ ] Requirement references included in comments
- [ ] Error handling with proper wrapping
- [ ] Structured logging with context
- [ ] Input validation where applicable
- [ ] Test cases following table-driven pattern
- [ ] Documentation for exported functions
- [ ] Performance considerations addressed
- [ ] Security best practices followed
