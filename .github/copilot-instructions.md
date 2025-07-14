<SYSTEM>
You are an AI programming GO senior developer/expert/assistant that is specialized in applying code changes to an existing document.
Follow Microsoft content policies.
Avoid content that violates copyrights.
If you are asked to generate content that is harmful, hateful, racist, sexist, lewd, violent, or completely irrelevant to software engineering, only respond with "Sorry, I can't assist with that."
Keep your answers short and impersonal.
The user has the following code open in the editor, starting from line 1.
</SYSTEM><|diff_marker|>

# GitHub Copilot Instructions for jonbu-ohlcv

## Project Overview
This is a Go-based OHLCV (Open, High, Low, Close, Volume) streaming service for stock market data that provides both historical data fetching and real-time streaming capabilities.

### Core Features
- **CLI Tool**: Fetch and display historical OHLCV data on demand
- **Real-time Streaming**: Ingest live market data from Alpaca API and other providers
- **WebSocket Server**: Stream live OHLCV candles to connected clients
- **REST API**: HTTP endpoints for data access and management
- **Database Storage**: PostgreSQL storage with proper schema design
- **Multi-Provider Support**: Pluggable data source architecture

### Use Cases
- Real-time market data streaming for trading applications
- Historical data analysis and backtesting
- Live charting and visualization
- Signal generation and algorithmic trading

---

## Technology Stack & Dependencies

### Core Technologies
- **Language**: Go 1.21+ (as specified in go.mod)
- **Database**: PostgreSQL with `lib/pq` driver
- **Web Framework**: Gorilla Mux for HTTP routing
- **Configuration**: Viper + godotenv for environment management
- **CLI Framework**: Cobra for command-line interface
- **Job Scheduling**: Robfig cron for periodic tasks

### Current Dependencies (from go.mod)
```go
// Core dependencies
github.com/gorilla/mux v1.8.1           // HTTP routing
github.com/joho/godotenv v1.5.1         // Environment file loading
github.com/lib/pq v1.10.9               // PostgreSQL driver
github.com/robfig/cron/v3 v3.0.1        // Job scheduling
github.com/rs/zerolog v1.34.0           // High-performance structured logging
github.com/spf13/cobra v1.8.0           // CLI framework
github.com/spf13/viper v1.17.0          // Configuration management
```

### Missing Dependencies (to be added when needed)
- `github.com/gorilla/websocket` - For WebSocket streaming
- `github.com/alpacahq/alpaca-trade-api-go` - Alpaca API client

### Additional Recommended Libraries
- `net/http` - Built-in HTTP server capabilities
- `go-cqrs/clock` or `time` - For precise interval control and time management

---

## Logging Strategy

**Current**: Using **zerolog** for high-performance structured logging

### Why zerolog?
- Extremely fast and low allocation  
- Structured JSON output by default, great for log aggregation systems  
- Simple API with leveled logging (debug, info, warn, error, fatal)  
- Easily integrates with context and fields for rich logs  

### Zerolog Setup
```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "os"
    "time"
)

func initLogger() {
    zerolog.TimeFieldFormat = time.RFC3339Nano
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano}).
        With().
        Timestamp().
        Logger()
}

func main() {
    initLogger()
    log.Info().Msg("Logger initialized")
}
```

### Logging Pattern
```go
// Service-level logging setup
type OHLCVService struct {
    logger zerolog.Logger
    // other fields...
}

func NewOHLCVService() *OHLCVService {
    logger := log.With().
        Str("service", "ohlcv").
        Str("component", "service").
        Logger()
    
    return &OHLCVService{
        logger: logger,
    }
}

// Method-level logging with context
func (s *OHLCVService) FetchOHLCV(symbol, timeframe string) error {
    logger := s.logger.With().
        Str("symbol", symbol).
        Str("timeframe", timeframe).
        Str("operation", "fetch").
        Logger()
    
    logger.Info().Msg("Starting OHLCV fetch")
    
    // Business logic here...
    if err != nil {
        logger.Error().Err(err).Msg("Failed to fetch OHLCV data")
        return err
    }
    
    logger.Info().
        Int("records", len(data)).
        Dur("duration", time.Since(start)).
        Msg("Successfully fetched OHLCV data")
    
    return nil
}

// Request-level logging with correlation IDs
func (h *Handler) GetOHLCV(w http.ResponseWriter, r *http.Request) {
    correlationID := uuid.New().String()
    logger := log.With().
        Str("correlation_id", correlationID).
        Str("method", r.Method).
        Str("path", r.URL.Path).
        Logger()
    
    logger.Info().Msg("Processing request")
    // Handle request...
}
```

---

## Go Coding Guidelines

### Core Principles
- **Idiomatic Go**: Follow `golang.org/doc/effective_go`
- **Simplicity**: Keep functions small, composable, and testable
- **Interfaces**: Use Go interfaces for abstraction where multiple implementations exist
- **Context**: Use context.Context for cancellation, timeouts, and request scoping
- **Error Handling**: Always handle errors explicitly with proper wrapping

### Code Organization
- Follow standard Go project layout (`cmd/`, `pkg/`, `internal/`)
- Use dependency injection for testability
- Separate concerns: transport layer, business logic, data access
- Keep exported APIs minimal and well-documented

### Naming Conventions
```go
// Exported types and functions: PascalCase
type OHLCVService struct{}
func (s *OHLCVService) GetHistoricalData() {}

// Unexported: camelCase
type internalConfig struct{}
func (c *internalConfig) loadFromFile() {}

// Interfaces: end with 'er' when possible
type DataFetcher interface {
    Fetch() error
}
```

---

## Project Structure (Current)
```
/home/ridopark/src/jonbu-ohlcv/
├── cmd/                    # Application entrypoints
│   ├── cli/               # CLI commands (jonbu-ohlcv cli)
│   ├── server/            # HTTP/WebSocket server
│   └── streamer/          # Alternative CLI entrypoint (streamer fetch)
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── database/         # Database operations and repositories
│   ├── fetcher/          # OHLCV data fetchers
│   │   └── alpaca/       # Alpaca API implementation
│   ├── models/           # Data models and structs (Candle, OHLCV, DTOs)
│   ├── service/          # Business logic layer
│   ├── scheduler/        # Job scheduling logic
│   ├── validator/        # Data validation
│   ├── streamsource/     # Stream source implementations
│   │   ├── alpaca/       # Alpaca WebSocket client
│   │   ├── polygon/      # Polygon API client (future)
│   │   └── mock/         # Mock for testing
│   ├── aggregator/       # Candle building from ticks/bars
│   ├── stream/           # WebSocket server logic
│   └── worker/           # Per-symbol worker processes
├── pkg/                   # Public libraries
│   └── api/              # API definitions
│       ├── handlers/     # HTTP handlers
│       ├── middleware/   # HTTP middleware
│       └── types/        # API request/response types
├── web/                   # Frontend React application (Phase 4)
│   ├── src/              # React TypeScript source code
│   ├── public/           # Static assets
│   ├── package.json      # Frontend dependencies
│   └── vite.config.ts    # Vite build configuration
├── migrations/            # Database schema migrations
├── config/               # Configuration files (.env.example)
├── docker/               # Docker configurations
├── scripts/              # Build and deployment scripts
├── test/                 # Test files and test data
└── .github/              # GitHub-specific files
    └── copilot-instructions.md
```

### Recommended Future Structure (for streaming)
```
internal/
├── streamsource/         # Stream source implementations
│   ├── alpaca/          # Alpaca WebSocket client
│   ├── polygon/         # Polygon API client
│   └── mock/            # Mock for testing
├── aggregator/          # Candle building from ticks/bars
├── stream/              # WebSocket server logic
└── worker/              # Per-symbol worker processes
```

*Note: Some of these directories are already reflected in the main structure above as the project evolves toward full streaming capabilities.*

---

## Web Frontend Architecture (Phase 4)

### Modern React + TypeScript Stack

The web frontend provides a comprehensive dashboard for OHLCV data visualization, CLI operations, and server monitoring with real-time capabilities.

#### Technology Stack
- **Framework**: React 18+ with TypeScript 5+ for type safety and modern React features
- **Styling**: Tailwind CSS 3+ with Less preprocessor for custom component styling
- **Build Tool**: Vite for fast development server and optimized production builds
- **Charts**: Recharts/Chart.js for real-time OHLCV candlestick charts and technical indicators
- **State Management**: Zustand for lightweight, TypeScript-friendly state management
- **WebSocket**: Socket.io client for real-time data streaming from Go backend
- **Routing**: React Router v6 for single-page application navigation

#### Project Structure
```
web/                       # Frontend application root
├── src/
│   ├── components/        # Reusable UI components
│   │   ├── charts/       # OHLCV chart components
│   │   │   ├── CandlestickChart.tsx
│   │   │   ├── TechnicalIndicators.tsx
│   │   │   └── VolumeChart.tsx
│   │   ├── forms/        # Form components for CLI operations
│   │   │   ├── SymbolForm.tsx
│   │   │   ├── FetchForm.tsx
│   │   │   └── MockControls.tsx
│   │   ├── layout/       # Layout and navigation
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Layout.tsx
│   │   └── ui/           # Basic UI components
│   │       ├── Button.tsx
│   │       ├── Input.tsx
│   │       └── Modal.tsx
│   ├── pages/            # Route-based page components
│   │   ├── Dashboard.tsx  # Main dashboard with charts
│   │   ├── Symbols.tsx    # Symbol management interface
│   │   ├── History.tsx    # Historical data viewer
│   │   ├── CLI.tsx        # Web-based CLI interface
│   │   ├── Monitoring.tsx # Server monitoring and health
│   │   └── Settings.tsx   # Configuration management
│   ├── hooks/            # Custom React hooks
│   │   ├── useWebSocket.ts # WebSocket connection management
│   │   ├── useOHLCV.ts    # OHLCV data fetching and streaming
│   │   ├── useAPI.ts      # REST API integration
│   │   └── useTheme.ts    # Dark/light theme management
│   ├── stores/           # Zustand state management
│   │   ├── chartStore.ts  # Chart data, settings, and indicators
│   │   ├── symbolStore.ts # Symbol tracking and management
│   │   ├── cliStore.ts    # CLI command history and state
│   │   └── configStore.ts # Application configuration
│   ├── services/         # API and WebSocket services
│   │   ├── api.ts         # REST API client with TypeScript types
│   │   ├── websocket.ts   # WebSocket client for real-time data
│   │   └── types.ts       # TypeScript type definitions for API
│   ├── styles/           # Styling and themes
│   │   ├── globals.less   # Global Less styles and variables
│   │   ├── components.less # Component-specific Less styles
│   │   ├── themes.less    # Dark/light theme definitions
│   │   └── tailwind.css   # Tailwind CSS imports and customizations
│   └── utils/            # Utility functions
│       ├── formatters.ts  # Data formatting utilities
│       ├── validators.ts  # Input validation helpers
│       └── constants.ts   # Application constants
├── public/               # Static assets
├── package.json         # Dependencies and scripts
├── tsconfig.json        # TypeScript configuration
├── tailwind.config.js   # Tailwind CSS configuration
├── vite.config.ts       # Vite build configuration
└── README.md            # Frontend documentation
```

### Frontend Features

#### Real-time Chart Dashboard
- **Candlestick Charts**: Live OHLCV data visualization with multiple timeframes
- **Technical Indicators**: SMA, EMA, RSI, MACD overlays with enriched candle data
- **Interactive Controls**: Zoom, pan, crosshair, timeframe selection
- **Volume Analysis**: Volume bars with market hours highlighting
- **Symbol Switching**: Quick symbol selection with favorites

#### Web-based CLI Interface
- **Command Execution**: All CLI commands available through web interface
- **Terminal Output**: Styled terminal display with command history
- **Auto-completion**: Smart command and parameter suggestions
- **Parameter Validation**: Real-time validation with helpful error messages
- **Command Palette**: Quick access to frequently used commands

#### Server Monitoring Dashboard
- **Health Status**: Real-time server component health indicators
- **Performance Metrics**: API response times, throughput, error rates
- **WebSocket Monitoring**: Active connections, subscription counts
- **Resource Usage**: CPU, memory, and database connection monitoring
- **Error Logs**: Filterable error log viewer with correlation IDs

#### Responsive Design System
- **Mobile-first**: Optimized for mobile, tablet, and desktop
- **Dark/Light Themes**: User-configurable theme with system preference detection
- **Component Library**: Consistent, reusable components with Tailwind + Less
- **Accessibility**: WCAG 2.1 compliance with keyboard navigation
- **Performance**: Code splitting, lazy loading, and optimized bundles

### Frontend Development Guidelines

#### React + TypeScript Best Practices
```typescript
// Component with proper TypeScript typing
interface ChartProps {
  symbol: string;
  timeframe: '1m' | '5m' | '15m' | '1h' | '1d';
  showIndicators: boolean;
  onTimeframeChange: (timeframe: string) => void;
}

const OHLCVChart: React.FC<ChartProps> = ({
  symbol,
  timeframe,
  showIndicators,
  onTimeframeChange
}) => {
  // Use custom hooks for data fetching
  const { candles, loading, error } = useOHLCV(symbol, timeframe);
  const { enrichedData } = useEnrichedCandles(symbol, timeframe);
  
  // Handle loading and error states
  if (loading) return <ChartSkeleton />;
  if (error) return <ErrorBoundary error={error} />;
  
  return (
    <div className="chart-container">
      <ChartHeader 
        symbol={symbol}
        timeframe={timeframe}
        onTimeframeChange={onTimeframeChange}
      />
      <CandlestickChart data={candles} />
      {showIndicators && (
        <TechnicalIndicators data={enrichedData} />
      )}
    </div>
  );
};
```

#### State Management with Zustand
```typescript
// Type-safe store with Zustand
interface ChartStore {
  symbol: string;
  timeframe: string;
  indicators: string[];
  theme: 'light' | 'dark';
  setSymbol: (symbol: string) => void;
  setTimeframe: (timeframe: string) => void;
  toggleIndicator: (indicator: string) => void;
  setTheme: (theme: 'light' | 'dark') => void;
}

const useChartStore = create<ChartStore>((set) => ({
  symbol: 'AAPL',
  timeframe: '1m',
  indicators: ['SMA20', 'EMA50'],
  theme: 'dark',
  setSymbol: (symbol) => set({ symbol }),
  setTimeframe: (timeframe) => set({ timeframe }),
  toggleIndicator: (indicator) => set((state) => ({
    indicators: state.indicators.includes(indicator)
      ? state.indicators.filter(i => i !== indicator)
      : [...state.indicators, indicator]
  })),
  setTheme: (theme) => set({ theme }),
}));
```

#### Tailwind + Less Integration
```less
// Component-specific styles with Tailwind integration
.chart-container {
  @apply bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4;
  
  .chart-header {
    @apply flex justify-between items-center mb-4;
    
    .symbol-title {
      @apply text-xl font-semibold text-gray-900 dark:text-white;
      
      // Custom Less styling
      &:hover {
        @apply text-blue-600 dark:text-blue-400;
        transition: color 0.2s ease;
      }
    }
    
    .timeframe-buttons {
      @apply flex space-x-2;
      
      button {
        @apply px-3 py-1 rounded-md text-sm font-medium;
        @apply bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600;
        
        &.active {
          @apply bg-blue-500 text-white hover:bg-blue-600;
        }
      }
    }
  }
}

// Responsive design with Less mixins
@media (max-width: 768px) {
  .chart-container {
    @apply p-2;
    
    .chart-header {
      @apply flex-col space-y-2;
    }
  }
}
```

#### WebSocket Integration
```typescript
// Real-time data streaming with proper error handling
const useWebSocket = (endpoint: string) => {
  const [socket, setSocket] = useState<Socket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const socketInstance = io(`ws://localhost:8080${endpoint}`);
    
    socketInstance.on('connect', () => {
      setIsConnected(true);
      setError(null);
    });
    
    socketInstance.on('disconnect', () => {
      setIsConnected(false);
    });
    
    socketInstance.on('error', (err) => {
      setError(err.message);
    });
    
    setSocket(socketInstance);
    
    return () => {
      socketInstance.disconnect();
    };
  }, [endpoint]);

  return { socket, isConnected, error };
};
```

---

## CLI Tool Implementation

### Command Structure
```bash
# Primary command structure (using cobra)
jonbu-ohlcv cli [command]

# Historical data fetching
jonbu-ohlcv cli fetch AAPL --timeframe 1d --start 2024-01-01 --end 2024-12-31
jonbu-ohlcv cli fetch GOOGL --output json

# Alternative streaming command pattern
jonbu-ohlcv streamer fetch AAPL --interval 1m --format table

# Symbol management
jonbu-ohlcv cli symbols add AAPL,GOOGL,MSFT
jonbu-ohlcv cli symbols list

# Database operations
jonbu-ohlcv cli migrate up
jonbu-ohlcv cli migrate down
```

### CLI Capabilities
- **Data Sources**: Connect to Alpaca API or load from local CSV files
- **Output Formats**: Support JSON, table, and CSV output formats
- **Symbol Management**: Add/remove symbols for tracking
- **Historical Fetching**: Fetch and display historical OHLCV data by symbol and timeframe
- **Real-time Preview**: Display live data streams for testing

### CLI Guidelines
- Use `cobra` for command structure (already in go.mod)
- Support multiple output formats with `--format` or `--output` flags
- Include progress indicators for long operations
- Provide clear error messages with actionable suggestions
- Support both interactive and batch modes
- Implement proper flag validation and help documentation

---

## Backend Server Architecture

### Core Responsibilities

#### Stream Ingestion
- Connect to Alpaca WebSocket or REST API
- Receive real-time tick/trade/quote data
- Aggregate raw data into OHLCV per timeframe (e.g., 1m, 5m, 15m, 1h)
- Handle connection failures and reconnection logic
- Rate limiting and backoff strategies

#### Candle Aggregation
- Implement a streaming `Aggregator` that emits completed OHLCV candles
- Support multiple timeframes and symbols simultaneously
- Handle interval rollovers and late-arriving ticks
- Aggregate ticks into OHLCV candles for multiple timeframes
- Support streaming aggregation (real-time)
- Handle late-arriving data and out-of-order events
- Emit completed candles via channels

#### API Server
```go
// REST endpoints using gorilla/mux (or chi as alternative)
GET    /api/v1/ohlcv/{symbol}                    // Latest candle
GET    /api/v1/ohlcv/{symbol}/history            // Historical data with query params
GET    /api/v1/symbols                           // Available symbols
POST   /api/v1/symbols                           // Add symbols to track
GET    /api/v1/market/status                     // Market status
GET    /api/v1/health                           // Health check endpoint
WebSocket /ws/ohlcv                              // Real-time streaming
```

#### WebSocket Streaming
- Use `gorilla/websocket` for real-time data streaming
- Support subscription by symbol and timeframe
- Handle client connections and disconnections gracefully
- Message queuing and backpressure handling
- Support subscribing by symbol and interval with real-time updates

---

## Architectural Patterns

### Core Design Principles
- Emit closed candles to `chan Candle` for downstream use
- Decouple stream producer (Alpaca) and consumer (WebSocket/API) via channels
- Handle reconnects and timeouts gracefully
- Use dependency injection for services

### Pluggable Streaming Ingestion

The ingestion of live market data must be modular and replaceable.

Define a `StreamSource` interface to abstract the data source. Each implementation connects to a live feed (e.g., Alpaca, Polygon) and pushes raw tick/trade data into a common format.

#### Example Interface

```go
type StreamSource interface {
    Connect(ctx context.Context) error
    Subscribe(symbols []string) error
    Read() <-chan MarketEvent  // emits ticks or trades
    Close() error
}
```
---

### 🧵 Per-Symbol Worker Processes

Each symbol is handled by a **dedicated worker process** (goroutine). This ensures that OHLCV data is streamed and aggregated **in parallel**, allowing for scalable and fault-isolated processing.

#### Responsibilities:
- Receive tick or bar data for a specific symbol
- Aggregate data into OHLCV candles (e.g., 1-minute)
- Emit completed candles through a channel
- Handle its own lifecycle and shutdown logic

> ⚠️ The Alpaca market data stream provides **1-minute OHLCV bars** by default.  
> To support custom timeframes (e.g., 5-minute or 15-minute), the worker should **accumulate and aggregate multiple 1-minute bars** into larger intervals.

#### Example Structure:

```go
type SymbolWorker struct {
    Symbol    string
    Ticks     chan MarketEvent    // Incoming events
    Candles   chan Candle         // Outgoing completed candles
    Quit      chan struct{}       // For graceful shutdown
}

func (w *SymbolWorker) Run() {
    aggregator := NewCandleAggregator(1 * time.Minute)
    for {
        select {
        case tick := <-w.Ticks:
            aggregator.AddTick(tick.Price, tick.Volume, tick.Timestamp)
        case candle := <-aggregator.CandleOut:
            w.Candles <- candle
        case <-w.Quit:
            return
        }
    }
}
```

## Best Practices:
- Each symbol gets its own goroutine
- Use buffered channels for Ticks and Candles
- Decouple upstream (data feed) and downstream (API/websocket) via channels
- Use a manager to start/stop workers per symbol dynamically

## Avoid
- No global state or shared mutable variables
- Avoid blocking channels without select statements
- Don’t mix business logic with transport (keep API separate from aggregation)
- Don’t store credentials or API keys in code
- Avoid processing multiple symbols in the same goroutine
- Don’t use shared mutable state between workers
- Avoid global dispatch logic inside the worker loop
- Do not rely on clock-based batching — drive aggregation from bar timestamps
- Avoid locking or shared state across workers
---

## Technology Stack Summary

### Key Technologies
- **Language**: Go 1.21+ (current project version)
- **Database**: PostgreSQL with `lib/pq` driver
- **Primary Data Provider**: Alpaca API for streaming and historical market data
- **Web Framework**: Gorilla Mux for HTTP routing
- **Configuration**: Environment variables with `godotenv` and `viper`
- **Logging**: Zerolog for high-performance structured logging
- **Scheduling**: Robfig cron for job scheduling
- **CLI**: Cobra for command-line interface

---

## Code Style & Conventions

### Go Best Practices
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- Use Go modules for dependency management
- Implement proper error handling with wrapped errors
- Use struct tags for JSON and database mapping: `json:"field_name" db:"field_name"`
- Include proper documentation comments for exported functions

### Database Patterns
- Use repository pattern for data access
- Implement proper transaction handling for batch operations
- Use prepared statements for performance
- Handle `sql.ErrNoRows` gracefully
- Include proper database connection management

### Error Handling
```go
// Always wrap errors with context
return fmt.Errorf("failed to insert OHLCV: %w", err)

// Handle sql.ErrNoRows specifically
if err == sql.ErrNoRows {
    return nil, nil
}
```

### Struct Definitions
```go
// Include proper JSON and database tags
type OHLCV struct {
    ID        int64     `json:"id" db:"id"`
    Symbol    string    `json:"symbol" db:"symbol"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

## Domain Knowledge

### OHLCV Data
- **Open**: Opening price for the time period
- **High**: Highest price during the time period
- **Low**: Lowest price during the time period
- **Close**: Closing price for the time period
- **Volume**: Total volume traded during the time period
- **Timeframes**: Common values include "1m", "5m", "15m", "1h", "4h", "1d"

### Stock Market Trading Symbols
- Format: Stock ticker symbols (e.g., "AAPL", "GOOGL", "MSFT")
- Exchange-specific symbols (e.g., "AAPL.O" for NASDAQ, "AAPL.US")
- Support for different markets (NYSE, NASDAQ, TSE, LSE, etc.)

### Data Provider Integration
- **Primary Provider**: Alpaca API for both streaming and historical data
- **Provider Abstraction**: Design interfaces to easily switch between different data providers
- **Multiple Provider Support**: Alpha Vantage, IEX Cloud, Polygon, Yahoo Finance, etc.
- Handle API rate limiting gracefully with provider-specific limits
- Consider market hours and trading sessions for accurate timestamps
- Handle different data formats and normalization across providers
- Implement fallback mechanisms when primary provider is unavailable

### Market Data Specifics
- **Market Hours**: Handle pre-market, regular hours, and after-hours trading sessions
- **Holidays**: Account for market holidays and closures
- **Splits & Dividends**: Consider stock splits and dividend adjustments
- **Corporate Actions**: Handle symbol changes, mergers, and delistings
- **Real-time vs Delayed**: Understand data delay requirements (15-20 minutes for most free APIs)
- **Market Maker Data**: Distinguish between bid/ask spread and last trade prices

## Development Guidelines

### When Adding New Features
1. **Models**: Add new structs to `internal/models/models.go` with proper tags
2. **Database**: Create repositories in `internal/database/` following existing patterns
3. **API**: Add endpoints to appropriate handlers in `pkg/api/`
4. **Configuration**: Add new config options to `internal/config/config.go`
5. **CLI**: Add new commands using Cobra in `cmd/cli/`
6. **Data Providers**: Add new provider implementations in `internal/fetcher/`
7. **Alpaca Integration**: Implement both streaming and historical data endpoints
8. **Provider Interfaces**: Ensure new providers implement common interfaces for easy switching

### Testing Considerations
- Write unit tests for business logic
- Use table-driven tests for multiple scenarios
- Mock database connections for testing
- Test error conditions thoroughly

### Database Operations
- Use transactions for batch operations
- Implement upsert operations with ON CONFLICT clauses
- Consider database constraints and indexes for performance
- Handle connection pooling and timeouts

### Enriched Candle Strategy: On-Demand Calculation
- **Storage Decision**: Store ONLY raw OHLCV data in PostgreSQL database
- **Enrichment Strategy**: Calculate enriched candles on-demand with 0.167ms latency
- **Rationale**: Real-time enrichment is faster than database I/O and avoids 50x storage bloat
- **Implementation**: Use enrichment engine for backtesting and analysis workflows
- **Benefits**: Flexible indicators, reduced storage costs, always current logic
- **Pattern**: Fetch raw OHLCV → Enrich on-demand → Return to client

```go
// Correct pattern: On-demand enrichment
func (s *BacktestService) GetEnrichedHistory(symbol string, from, to time.Time) ([]*models.EnrichedCandle, error) {
    // 1. Fetch raw OHLCV from database (lightweight)
    ohlcv, err := s.repo.GetOHLCVHistory(symbol, from, to)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
    }
    
    // 2. Enrich on-demand (0.167ms per candle)
    enriched := make([]*models.EnrichedCandle, len(ohlcv))
    for i, candle := range ohlcv {
        enriched[i], err = s.enricher.EnrichCandle(ctx, candle, ohlcv[:i], nil)
        if err != nil {
            return nil, fmt.Errorf("enrichment failed: %w", err)
        }
    }
    
    return enriched, nil
}
```

### API Design
- Follow RESTful conventions
- Use proper HTTP status codes
- Implement pagination for large datasets
- Include proper content-type headers
- Handle query parameters for filtering

### Configuration Management
- Use environment variables with sensible defaults
- Support .env files for development
- Validate configuration on startup
- Document all configuration options

### Logging & Monitoring
- Use structured logging with zerolog
- Include relevant context in log messages
- Log at appropriate levels (debug, info, warn, error)
- Include request IDs for tracing

## Common Patterns

### Data Provider Interface Pattern
```go
// Define common interface for all data providers
type DataProvider interface {
    GetHistoricalOHLCV(symbol, timeframe string, start, end time.Time) ([]*models.OHLCV, error)
    StreamOHLCV(symbols []string, timeframe string) (<-chan *models.OHLCV, error)
    ValidateSymbol(symbol string) error
    GetMarketStatus() (*MarketStatus, error)
}

// Alpaca-specific implementation
type AlpacaProvider struct {
    apiKey    string
    secretKey string
    baseURL   string
}

func (a *AlpacaProvider) GetHistoricalOHLCV(symbol, timeframe string, start, end time.Time) ([]*models.OHLCV, error) {
    // Alpaca API implementation
}
```

### Provider Factory Pattern
```go
func NewDataProvider(providerType string, config map[string]string) (DataProvider, error) {
    switch providerType {
    case "alpaca":
        return NewAlpacaProvider(config), nil
    case "polygon":
        return NewPolygonProvider(config), nil
    default:
        return nil, fmt.Errorf("unsupported provider: %s", providerType)
    }
}
```

### Repository Pattern
```go
type OHLCVRepository struct {
    db *DB
}

func NewOHLCVRepository(db *DB) *OHLCVRepository {
    return &OHLCVRepository{db: db}
}

func (r *OHLCVRepository) Insert(ohlcv *models.OHLCV) error {
    // Implementation with proper error handling
}
```

### Configuration Loading
```go
func Load() (*Config, error) {
    // Load .env file if exists
    if err := godotenv.Load("config/.env"); err != nil {
        log.Warn().Msg("No .env file found, using environment variables")
    }
    // Continue with environment variable parsing
}
```

### HTTP Handler Pattern
```go
func (h *Handler) GetOHLCV(w http.ResponseWriter, r *http.Request) {
    // Extract parameters
    // Validate input
    // Call business logic
    // Return JSON response
}
```

## Security & Deployment

### Security Best Practices
```go
// Input validation for API endpoints
func validateSymbol(symbol string) error {
    if len(symbol) == 0 || len(symbol) > 10 {
        return errors.New("symbol must be 1-10 characters")
    }
    if !regexp.MustCompile(`^[A-Z]+$`).MatchString(symbol) {
        return errors.New("symbol must contain only uppercase letters")
    }
    return nil
}

// Rate limiting middleware
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Implement rate limiting logic
        next.ServeHTTP(w, r)
    })
}
```

### Secret Management
- Store API keys in environment variables, never in code
- Use separate credentials for development, staging, and production
- Rotate API keys regularly
- Implement proper access controls for database credentials

### Production Deployment
- Use HTTPS for all API endpoints
- Implement proper CORS policies
- Set up graceful shutdown handling
- Configure proper timeouts and circuit breakers
- Use connection pooling for database connections

### Docker Configuration
```dockerfile
# Multi-stage build for optimized production image
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

---

## Troubleshooting & Common Pitfalls

### Common Issues

#### Database Connection Problems
```go
// Always check database connectivity on startup
func validateDBConnection(db *sql.DB) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("database connection failed: %w", err)
    }
    return nil
}
```

#### API Rate Limiting
```go
// Implement exponential backoff for API failures
func retryWithBackoff(operation func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := operation(); err == nil {
            return nil
        }
        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(backoff)
    }
    return errors.New("max retries exceeded")
}
```

#### Memory Leaks in Streaming
- Always close channels and goroutines properly
- Use context for cancellation in long-running operations
- Monitor goroutine count in production
- Implement proper cleanup in defer statements

#### Time Zone Handling
```go
// Always use market timezone for OHLCV timestamps
func parseMarketTime(timeStr string) (time.Time, error) {
    loc, err := time.LoadLocation("America/New_York")
    if err != nil {
        return time.Time{}, err
    }
    return time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc)
}
```

### Performance Optimization
- Use prepared statements for repeated database queries
- Implement connection pooling with appropriate limits
- Cache frequently accessed data with TTL
- Use buffered channels to prevent goroutine blocking
- Profile memory usage regularly with `go tool pprof`

### Debugging Tips
- Use structured logging with correlation IDs
- Implement detailed error messages with context
- Use Go's built-in race detector: `go run -race`
- Monitor resource usage with `/debug/pprof` endpoints
- Test with market data edge cases (holidays, split dates)

---

## Development Checklist

When implementing new features, ensure:

- [ ] Proper error handling with context
- [ ] Unit tests with table-driven patterns
- [ ] Input validation for all public APIs
- [ ] Structured logging with appropriate levels
- [ ] Documentation for exported functions
- [ ] Database migrations if schema changes
- [ ] Rate limiting for external API calls
- [ ] Graceful shutdown handling
- [ ] Memory leak prevention (close channels/goroutines)
- [ ] Time zone awareness for market data
- [ ] Configuration validation on startup
- [ ] Health check endpoints for monitoring

---

*This document should be updated as the project evolves. When adding new patterns, libraries, or architectural decisions, update the relevant sections to maintain consistency and best practices.*

**When suggesting code improvements or new features, always consider these patterns and maintain consistency with the existing codebase structure.**

## Event-Driven Pipeline with Aggregator Workers

### Architecture Overview

The OHLCV streaming service uses an event-driven pipeline architecture where **aggregator workers** are spawned per symbol and interval combination. This design ensures high throughput, scalability, and clean separation of concerns.

### Worker Lifecycle

#### Worker Spawning Strategy
```go
type AggregatorManager struct {
    workers map[string]*AggregatorWorker
    symbols []string
    intervals []time.Duration
    inputChan chan MarketEvent
    outputChan chan Candle
}

func (m *AggregatorManager) Start() {
    for _, symbol := range m.symbols {
        for _, interval := range m.intervals {
            worker := NewAggregatorWorker(symbol, interval)
            go worker.Run(m.inputChan, m.outputChan)
            m.workers[fmt.Sprintf("%s:%s", symbol, interval)] = worker
        }
    }
}
```

#### Individual Worker Implementation
```go
type AggregatorWorker struct {
    Symbol       string
    Interval     time.Duration
    currentCandle *Candle
    buffer       []MarketEvent
    lastEmit     time.Time
}

func (w *AggregatorWorker) Run(input <-chan MarketEvent, output chan<- Candle) {
    for {
        select {
        case event := <-input:
            if event.Symbol != w.Symbol {
                continue // Skip events for other symbols
            }
            
            w.processEvent(event)
            
            // Check if interval is complete
            if w.shouldEmitCandle(event.Timestamp) {
                if w.currentCandle != nil {
                    output <- *w.currentCandle
                    log.Info().
                        Str("symbol", w.Symbol).
                        Dur("interval", w.Interval).
                        Time("timestamp", w.currentCandle.Timestamp).
                        Msg("Emitted aggregated candle")
                }
                w.startNewCandle(event.Timestamp)
            }
        }
    }
}
```

### Incremental Aggregation Logic

#### 1-Minute Base Aggregation
```go
func (w *AggregatorWorker) processEvent(event MarketEvent) {
    if w.currentCandle == nil {
        w.startNewCandle(event.Timestamp)
    }
    
    // Update OHLCV values incrementally
    if w.currentCandle.Open == 0 {
        w.currentCandle.Open = event.Price
    }
    
    if event.Price > w.currentCandle.High {
        w.currentCandle.High = event.Price
    }
    
    if event.Price < w.currentCandle.Low || w.currentCandle.Low == 0 {
        w.currentCandle.Low = event.Price
    }
    
    w.currentCandle.Close = event.Price
    w.currentCandle.Volume += event.Volume
    w.currentCandle.LastUpdate = event.Timestamp
}
```

#### Multi-Interval Aggregation
```go
// For 5-minute intervals, aggregate from 1-minute bars
func (w *AggregatorWorker) processOneMinuteBar(bar Candle) {
    if w.Interval == time.Minute {
        // Direct 1-minute processing
        w.emitCandle(bar)
        return
    }
    
    // Aggregate multiple 1-minute bars for larger intervals
    w.buffer = append(w.buffer, bar)
    
    if w.shouldEmitAggregatedCandle() {
        aggregated := w.aggregateBuffer()
        w.emitCandle(aggregated)
        w.clearBuffer()
    }
}

func (w *AggregatorWorker) aggregateBuffer() Candle {
    if len(w.buffer) == 0 {
        return Candle{}
    }
    
    result := Candle{
        Symbol:    w.Symbol,
        Interval:  w.Interval,
        Timestamp: w.buffer[0].Timestamp,
        Open:      w.buffer[0].Open,
        High:      w.buffer[0].High,
        Low:       w.buffer[0].Low,
        Close:     w.buffer[len(w.buffer)-1].Close,
        Volume:    0,
    }
    
    for _, bar := range w.buffer {
        if bar.High > result.High {
            result.High = bar.High
        }
        if bar.Low < result.Low {
            result.Low = bar.Low
        }
        result.Volume += bar.Volume
    }
    
    return result
}
```

### Channel-Based Communication

#### Pipeline Architecture
```go
type StreamingPipeline struct {
    // Input from data sources (Alpaca, etc.)
    rawEvents     chan MarketEvent
    
    // Processed 1-minute candles
    minuteCandles chan Candle
    
    // Multi-interval aggregated candles
    aggregatedCandles chan Candle
    
    // Output to WebSocket clients
    clientOutputs map[string]chan Candle
}

func (p *StreamingPipeline) Start() {
    // Start base 1-minute aggregators
    go p.runBaseAggregators()
    
    // Start multi-interval aggregators
    go p.runIntervalAggregators()
    
    // Start client distributors
    go p.distributeToClients()
}
```

#### Backpressure Handling
```go
func (w *AggregatorWorker) emitWithBackpressure(candle Candle, output chan<- Candle) {
    select {
    case output <- candle:
        // Successfully sent
    case <-time.After(100 * time.Millisecond):
        // Channel is full, log warning but don't block
        log.Warn().
            Str("symbol", candle.Symbol).
            Msg("Output channel full, dropping candle")
    }
}
```

### Advantages

#### ✅ Pros
- **High Parallelism**: Each symbol-interval combination runs independently
- **Easy Scaling**: Add workers dynamically as new symbols are tracked
- **Clean Separation**: Aggregation logic isolated from transport layer
- **Fault Isolation**: Worker failure doesn't affect other symbols
- **Memory Efficiency**: Each worker only holds data for its symbol
- **Real-time Processing**: No batching delays, immediate aggregation

#### Resource Management
```go
type WorkerPool struct {
    workers     map[string]*AggregatorWorker
    maxWorkers  int
    activeCount int
    mu          sync.RWMutex
}

func (p *WorkerPool) AddSymbol(symbol string, intervals []time.Duration) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.activeCount+len(intervals) > p.maxWorkers {
        return fmt.Errorf("worker pool at capacity")
    }
    
    for _, interval := range intervals {
        worker := NewAggregatorWorker(symbol, interval)
        key := fmt.Sprintf("%s:%s", symbol, interval)
        p.workers[key] = worker
        go worker.Run()
        p.activeCount++
    }
    
    return nil
}
```

### Challenges & Solutions

#### ⚠️ Cons & Mitigation Strategies

**Goroutine Management Complexity**
```go
// Solution: Centralized worker lifecycle management
type WorkerManager struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func (m *WorkerManager) Shutdown() {
    m.cancel() // Signal all workers to stop
    m.wg.Wait() // Wait for graceful shutdown
}
```

**Channel Buffering & Memory**
```go
// Solution: Adaptive buffer sizing based on throughput
func calculateBufferSize(symbol string, interval time.Duration) int {
    baseSize := 100
    
    // Adjust based on symbol volatility and interval
    if interval < time.Minute {
        baseSize *= 5 // High-frequency needs larger buffer
    }
    
    return baseSize
}
```

**Out-of-Order Event Handling**
```go
// Solution: Time-based ordering within worker
type OrderedBuffer struct {
    events []MarketEvent
    maxAge time.Duration
}

func (b *OrderedBuffer) Add(event MarketEvent) []MarketEvent {
    b.events = append(b.events, event)
    
    // Sort by timestamp
    sort.Slice(b.events, func(i, j int) bool {
        return b.events[i].Timestamp.Before(b.events[j].Timestamp)
    })
    
    // Emit events older than maxAge
    cutoff := time.Now().Add(-b.maxAge)
    ready := []MarketEvent{}
    
    for i, event := range b.events {
        if event.Timestamp.Before(cutoff) {
            ready = append(ready, event)
        } else {
            b.events = b.events[i:]
            break
        }
    }
    
    return ready
}
```

### Performance Considerations

- **Buffer Sizing**: Start with 100-1000 element buffers, adjust based on throughput
- **Worker Count**: Typically 1-5 workers per symbol depending on intervals tracked
- **Memory Usage**: ~1-10MB per worker depending on buffer size and candle history
- **Latency**: Sub-millisecond aggregation latency for real-time streaming
- **Throughput**: 10k+ events/second per worker on modern hardware

---

## Requirements & Specifications

This project follows detailed requirements documented in `.github/copilot-requirements.md`. When generating code, ensure compliance with:

### Critical Requirements Categories
- **Data Ingestion**: Real-time processing, multi-provider support, connection handling
- **Performance**: 10k+ events/second, sub-millisecond latency, memory efficiency
- **Reliability**: Health checks, graceful degradation, error recovery
- **Security**: Input validation, SQL injection prevention, secret management
- **Code Quality**: >80% test coverage, structured logging, documentation

### Key Implementation Rules
1. **Always validate user input** before processing
2. **Use structured logging** with correlation IDs for all operations
3. **Implement proper error wrapping** with context for debugging
4. **Follow table-driven test patterns** for comprehensive coverage
5. **Use dependency injection** for testability and modularity

### Code Generation Guidelines
Refer to `.github/copilot-generation-guide.md` for:
- Requirement-driven development patterns
- Mandatory code templates and examples
- Architecture compliance patterns
- Performance and monitoring integration

### Compliance Checks
Before implementing any feature, verify:
- [ ] Requirements traceability (REQ-XXX references)
- [ ] Error handling and logging implementation
- [ ] Test coverage for new functionality
- [ ] Documentation updates
- [ ] Performance impact assessment

---
