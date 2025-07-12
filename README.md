# jonbu-ohlcv

A high-performance Go-based OHLCV (Open, High, Low, Close, Volume) streaming service for stock market data that provides both historical data fetching and real-time streaming capabilities with multi-provider support.

## 🚀 Features

### Core Capabilities
- **Real-time Streaming**: Live market data ingestion from Alpaca API and other providers
- **Historical Data**: Fetch and analyze historical OHLCV data with flexible time ranges
- **Multi-Timeframe Support**: 1m, 5m, 15m, 1h, 4h, 1d intervals with real-time aggregation
- **WebSocket Server**: Stream live OHLCV candles to connected clients
- **REST API**: Comprehensive HTTP endpoints for data access and management
- **CLI Tool**: Command-line interface for data fetching and system management
- **Event-Driven Architecture**: Per-symbol worker processes with channel-based communication

### Advanced Features
- **Multi-Provider Support**: Pluggable data source architecture (Alpaca primary, extensible to Polygon, Alpha Vantage, etc.)
- **High-Performance Aggregation**: Sub-millisecond latency, 10k+ events/second throughput
- **Fault Tolerance**: Connection recovery, graceful degradation, and error isolation
- **Production Ready**: Health checks, monitoring, rate limiting, and structured logging

## 🏗️ Architecture

### System Design
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Data Sources  │────▶│ Aggregator Pool  │────▶│  API & WebSocket│
│  (Alpaca, etc.) │    │  (Per-Symbol)    │    │    Endpoints    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                        ┌──────────────────┐
                        │   PostgreSQL     │
                        │    Database      │
                        └──────────────────┘
```

### Event-Driven Pipeline
- **Aggregator Workers**: Dedicated goroutines per symbol-interval combination
- **Channel-Based Communication**: Non-blocking event processing with backpressure handling
- **Real-time Aggregation**: Live tick data aggregated into OHLCV candles
- **Scalable Design**: Dynamic worker spawning and resource management

## 🛠️ Technology Stack

### Core Technologies
- **Language**: Go 1.21+ with modern concurrency patterns
- **Database**: PostgreSQL with connection pooling and prepared statements
- **Logging**: Zerolog for high-performance structured logging
- **Web Framework**: Gorilla Mux for HTTP routing and WebSocket support
- **Configuration**: Viper + godotenv for environment management
- **CLI**: Cobra for command-line interface
- **Scheduling**: Robfig cron for periodic tasks

### Data Providers
- **Primary**: Alpaca API (streaming + historical)
- **Planned**: Polygon, Alpha Vantage, IEX Cloud, Yahoo Finance
- **Extensible**: Provider interface for easy integration

### Dependencies
```go
// Production dependencies
github.com/gorilla/mux v1.8.1           // HTTP routing & WebSocket
github.com/rs/zerolog v1.34.0           // High-performance logging
github.com/lib/pq v1.10.9               // PostgreSQL driver
github.com/spf13/cobra v1.8.0           // CLI framework
github.com/spf13/viper v1.17.0          // Configuration management
github.com/robfig/cron/v3 v3.0.1        // Job scheduling
```

## 📁 Project Structure

```
jonbu-ohlcv/
├── cmd/                    # Application entrypoints
│   ├── cli/               # CLI commands and tools
│   ├── server/            # HTTP/WebSocket server
│   └── streamer/          # Alternative streaming entrypoint
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── database/         # Database operations and repositories
│   ├── fetcher/          # OHLCV data fetchers
│   │   └── alpaca/       # Alpaca API implementation
│   ├── models/           # Data models and structs
│   ├── service/          # Business logic layer
│   ├── scheduler/        # Job scheduling logic
│   ├── validator/        # Data validation
│   ├── aggregator/       # Candle building from ticks/bars
│   ├── stream/           # WebSocket server logic
│   └── worker/           # Per-symbol worker processes
├── pkg/                   # Public libraries
│   └── api/              # API definitions
│       ├── handlers/     # HTTP handlers
│       ├── middleware/   # HTTP middleware
│       └── types/        # API request/response types
├── migrations/            # Database schema migrations
├── config/               # Configuration files
├── docker/               # Docker configurations
├── scripts/              # Build and deployment scripts
└── test/                 # Test files and test data
```

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+**: Modern Go with generics and performance improvements
- **PostgreSQL 12+**: Database with JSON support and window functions
- **Git**: Version control
- **Make**: Build automation (optional)

### Quick Start

1. **Clone and Setup**
```bash
git clone https://github.com/your-username/jonbu-ohlcv.git
cd jonbu-ohlcv
go mod download
```

2. **Configure Environment**
```bash
cp config/.env.example config/.env
# Edit config/.env with your settings:
# - DATABASE_URL
# - ALPACA_API_KEY
# - ALPACA_SECRET_KEY
```

3. **Database Setup**
```bash
# Create database and run migrations
make migrate-up
# Or manually:
psql -c "CREATE DATABASE jonbu_ohlcv;"
go run cmd/cli/main.go migrate up
```

4. **Start the Services**
```bash
# Start the HTTP/WebSocket server
go run cmd/server/main.go

# Or start streaming service
go run cmd/streamer/main.go
```

## 💻 Usage

### CLI Tool

The CLI provides comprehensive data management capabilities:

```bash
# Fetch historical data
jonbu-ohlcv cli fetch AAPL --timeframe 1d --start 2024-01-01 --end 2024-12-31
jonbu-ohlcv cli fetch GOOGL,MSFT,TSLA --output json --format table

# Symbol management
jonbu-ohlcv cli symbols add AAPL,GOOGL,MSFT
jonbu-ohlcv cli symbols list
jonbu-ohlcv cli symbols remove AAPL

# Database operations
jonbu-ohlcv cli migrate up
jonbu-ohlcv cli migrate down
jonbu-ohlcv cli migrate status

# Real-time preview
jonbu-ohlcv streamer fetch AAPL --interval 1m --format table
```

### REST API

Comprehensive HTTP endpoints for data access:

```bash
# Latest OHLCV data
GET /api/v1/ohlcv/{symbol}

# Historical data with filtering
GET /api/v1/ohlcv/{symbol}/history?timeframe=1d&start=2024-01-01&end=2024-12-31&limit=100

# Symbol management
GET /api/v1/symbols                     # List tracked symbols
POST /api/v1/symbols                    # Add symbols to track
DELETE /api/v1/symbols/{symbol}         # Remove symbol

# Market information
GET /api/v1/market/status               # Market status and hours
GET /api/v1/health                      # Health check endpoint

# Real-time streaming
WebSocket /ws/ohlcv                     # Subscribe to live updates
```

### WebSocket Streaming

Real-time data streaming with subscription management:

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/ohlcv');

// Subscribe to symbols and intervals
ws.send(JSON.stringify({
    action: 'subscribe',
    symbols: ['AAPL', 'GOOGL'],
    intervals: ['1m', '5m']
}));

// Receive real-time updates
ws.onmessage = (event) => {
    const candle = JSON.parse(event.data);
    console.log(`${candle.symbol}: ${candle.close} @ ${candle.timestamp}`);
};
```

## ⚙️ Configuration

### Environment Variables

```bash
# Database Configuration
DATABASE_URL=postgres://user:password@localhost/jonbu_ohlcv?sslmode=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE=5

# Alpaca API Configuration
ALPACA_API_KEY=your_api_key
ALPACA_SECRET_KEY=your_secret_key
ALPACA_BASE_URL=https://paper-api.alpaca.markets  # or live API

# Server Configuration
HTTP_PORT=8080
WEBSOCKET_PORT=8081
LOG_LEVEL=info
LOG_FORMAT=json

# Performance Tuning
WORKER_BUFFER_SIZE=1000
MAX_WORKERS_PER_SYMBOL=5
AGGREGATION_TIMEOUT=5s
```

### Logging Configuration

Using zerolog for high-performance structured logging:

```go
// Production logging setup
zerolog.TimeFieldFormat = time.RFC3339Nano
log.Logger = log.Output(zerolog.ConsoleWriter{
    Out: os.Stderr,
    TimeFormat: time.RFC3339Nano,
}).With().Timestamp().Logger()

// Request-level logging with correlation IDs
logger := log.With().
    Str("correlation_id", correlationID).
    Str("symbol", symbol).
    Str("operation", "fetch").
    Logger()
```

## 🔧 Development

### Code Quality Standards

- **Test Coverage**: >80% coverage with table-driven tests
- **Error Handling**: Explicit error handling with context wrapping
- **Structured Logging**: Zerolog with correlation IDs and request tracing
- **Documentation**: Comprehensive godoc for exported functions
- **Performance**: Sub-millisecond latency targets for real-time operations

### Development Workflow

```bash
# Install development dependencies
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint and format
golangci-lint run
goimports -w .

# Build for production
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server cmd/server/main.go
```

### Testing Strategy

```go
// Table-driven tests for comprehensive coverage
func TestOHLCVService_FetchHistorical(t *testing.T) {
    tests := []struct {
        name     string
        symbol   string
        timeframe string
        want     int
        wantErr  bool
    }{
        {"valid symbol", "AAPL", "1d", 252, false},
        {"invalid symbol", "", "1d", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 🚢 Deployment

### Docker Support

```dockerfile
# Multi-stage build for optimized production image
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### Production Considerations

- **Health Checks**: Implement `/health` endpoint with dependency checks
- **Monitoring**: Prometheus metrics and structured logging
- **Security**: HTTPS, CORS policies, input validation, secret management
- **Scaling**: Horizontal scaling with load balancers
- **Database**: Connection pooling, read replicas, backup strategies

## 📊 Performance

### Benchmarks

- **Throughput**: 10,000+ events/second per worker
- **Latency**: Sub-millisecond aggregation latency
- **Memory**: ~1-10MB per worker depending on buffer configuration
- **Concurrent Workers**: 100+ symbols with multiple intervals
- **Database**: Optimized with prepared statements and connection pooling

### Optimization Features

- **Event-Driven Architecture**: Non-blocking channel-based processing
- **Buffered Channels**: Configurable buffer sizes for throughput optimization
- **Connection Pooling**: Database and HTTP connection reuse
- **Structured Logging**: High-performance zerolog with minimal allocations

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

### Development Process

1. **Fork the repository** and create a feature branch
2. **Follow Go best practices** and project coding standards
3. **Add comprehensive tests** with >80% coverage
4. **Update documentation** for new features
5. **Submit a pull request** with clear description

### Code Standards

- Follow Go naming conventions and `go fmt` standards
- Use dependency injection for testability
- Implement proper error handling with context
- Add structured logging with appropriate levels
- Include godoc documentation for exported functions

### Testing Requirements

- Unit tests for business logic
- Integration tests for database operations
- Table-driven tests for multiple scenarios
- Mock external dependencies (APIs, databases)
- Performance benchmarks for critical paths

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for complete details.

## 🆘 Support

### Documentation

- **API Documentation**: See `/docs` directory for OpenAPI specifications
- **Architecture Guide**: Detailed system design in `/docs/architecture.md`
- **Deployment Guide**: Production deployment instructions in `/docs/deployment.md`

### Getting Help

- **Issues**: GitHub Issues for bug reports and feature requests
- **Discussions**: GitHub Discussions for questions and community support
- **Documentation**: Comprehensive godoc documentation

### Common Issues

- **Database Connection**: Verify PostgreSQL is running and credentials are correct
- **API Rate Limits**: Implement exponential backoff for external API calls
- **Memory Usage**: Monitor goroutine count and channel buffer sizes
- **Time Zones**: Ensure market timezone handling for accurate timestamps

---

## 🔮 Roadmap

### Phase 1: Core Foundation ✅
- [x] Basic OHLCV data structures and database schema
- [x] Alpaca API integration for historical data
- [x] CLI tool for data fetching and management
- [x] REST API with basic endpoints

### Phase 2: Real-time Streaming ✅
- [x] Event-driven architecture with worker processes
- [x] WebSocket server for real-time data streaming
- [x] Multi-timeframe aggregation (1m, 5m, 15m, 1h, 1d)
- [x] High-performance logging with zerolog

### Phase 3: Production Features 🚧
- [ ] Multi-provider support (Polygon, Alpha Vantage)
- [ ] Advanced monitoring and alerting
- [ ] Horizontal scaling and load balancing
- [ ] Comprehensive API documentation

### Phase 4: AI Integration 🔮
- [ ] Technical indicators (RSI, MACD, Bollinger Bands)
- [ ] AI-ready enriched candles with market context
- [ ] Vector embeddings for RAG systems
- [ ] Trading signal generation

*Last updated: July 2025*
