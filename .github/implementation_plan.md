# Implementation Plan for jonbu-ohlcv

## Overview

This implementation plan addresses all 230+ requirements across 4 phases, from core infrastructure to AI integration. Each phase builds upon the previous, ensuring a solid foundation for advanced features.

## 📋 Implementation Phases

### Phase 1: Core Infrastructure 🏗️
**Target Timeline**: 4-6 weeks  
**Requirements**: REQ-001 to REQ-100  
**Success Criteria**: Basic OHLCV ingestion, storage, and API access

### Phase 2: Real-time Streaming 🚀
**Target Timeline**: 3-4 weeks  
**Requirements**: REQ-031 to REQ-040, REQ-017 to REQ-020  
**Success Criteria**: WebSocket streaming with 10k+ events/second

### Phase 3: AI Integration 🧠
**Target Timeline**: 4-5 weeks  
**Requirements**: REQ-200 to REQ-220  
**Success Criteria**: Real-time technical indicators and market analysis

### Phase 4: RAG Integration 🤖
**Target Timeline**: 3-4 weeks  
**Requirements**: REQ-221 to REQ-230  
**Success Criteria**: Vector embeddings and RAG-ready data pipeline

### Phase 5: Production Hardening 🛡️
**Target Timeline**: 2-3 weeks  
**Requirements**: REQ-081 to REQ-100, Security & Monitoring  
**Success Criteria**: Production-ready deployment with 99.9% uptime

---

## Phase 1: Core Infrastructure Implementation

### 1.1 Project Foundation (Week 1)

#### Directory Structure Setup
```bash
internal/
├── config/          # REQ-061 to REQ-065
│   ├── config.go
│   ├── validator.go
│   └── env.go
├── models/          # REQ-071 to REQ-075
│   ├── ohlcv.go
│   ├── market_event.go
│   └── candle.go
├── database/        # REQ-011 to REQ-015
│   ├── connection.go
│   ├── repository.go
│   └── migrations/
└── logger/          # REQ-046 to REQ-050
    ├── logger.go
    └── middleware.go
```

**Key Components:**
- **Config Management** (REQ-061, REQ-062, REQ-065): Viper-based configuration with validation
- **Data Models** (REQ-071 to REQ-075): OHLCV structures with proper tags
- **Database Layer** (REQ-011 to REQ-015): PostgreSQL with connection pooling
- **Logging Foundation** (REQ-046 to REQ-050): Zerolog with correlation IDs

### 1.2 Data Ingestion Layer (Week 2)

#### Alpaca Integration
```go
// REQ-001, REQ-003, REQ-005
internal/fetcher/alpaca/
├── client.go        # HTTP client for historical data
├── websocket.go     # WebSocket client for real-time
├── types.go         # Alpaca-specific data structures
└── validator.go     # Data validation and normalization
```

**Implementation Focus:**
- **REQ-001**: Real-time WebSocket connection to Alpaca
- **REQ-003**: Automatic reconnection with exponential backoff
- **REQ-004**: Rate limiting and circuit breaker patterns
- **REQ-005**: Data validation and integrity checks
- **REQ-029**: Provider interface for future extensibility

### 1.3 Data Processing Core (Week 2-3)

#### Aggregation Engine
```go
// REQ-006 to REQ-010
internal/aggregator/
├── worker.go        # Per-symbol aggregation worker
├── manager.go       # Worker lifecycle management
├── candle.go        # OHLCV candle building logic
└── timeframe.go     # Multi-timeframe support
```

**Key Features:**
- **REQ-006**: Multi-timeframe aggregation (1m, 5m, 15m, 1h, 1d)
- **REQ-007**: Out-of-order event handling with time-based sorting
- **REQ-008**: Incremental real-time aggregation
- **REQ-009**: Per-symbol worker processes for parallelism
- **REQ-010**: Channel-based candle emission

### 1.4 Database Operations (Week 3)

#### Repository Pattern
```go
// REQ-011 to REQ-015
internal/database/
├── connection.go    # Connection pooling and management
├── ohlcv_repo.go   # OHLCV data operations
├── symbol_repo.go  # Symbol management
└── migrations/     # Schema versioning
    ├── 001_initial.sql
    ├── 002_indexes.sql
    └── 003_constraints.sql
```

**Database Design:**
- **REQ-011**: PostgreSQL with optimized schema
- **REQ-012**: Migration system for schema evolution
- **REQ-013**: Prepared statements for performance
- **REQ-014**: Connection pooling with configurable limits
- **REQ-015**: Transaction handling for batch operations

### 1.5 REST API Foundation (Week 4)

#### API Layer
```go
// REQ-016, REQ-041 to REQ-045
pkg/api/
├── handlers/
│   ├── ohlcv.go     # OHLCV endpoints
│   ├── symbols.go   # Symbol management
│   └── health.go    # Health checks
├── middleware/
│   ├── auth.go      # Authentication
│   ├── logging.go   # Request logging
│   └── ratelimit.go # Rate limiting
└── types/
    ├── request.go   # Request DTOs
    └── response.go  # Response DTOs
```

**API Endpoints:**
- `GET /api/v1/ohlcv/{symbol}` - Latest candle data
- `GET /api/v1/ohlcv/{symbol}/history` - Historical data with filtering
- `GET /api/v1/symbols` - Symbol management
- `POST /api/v1/symbols` - Add tracking symbols
- `GET /api/v1/health` - Health status

### 1.6 CLI Tool (Week 4)

#### Command Structure
```go
// REQ-021 to REQ-025
cmd/cli/
├── main.go
├── fetch.go         # Data fetching commands
├── symbols.go       # Symbol management
├── migrate.go       # Database migrations
└── server.go        # Server management
```

**CLI Commands:**
```bash
# Data operations
jonbu-ohlcv cli fetch AAPL --timeframe 1d --start 2024-01-01
jonbu-ohlcv cli symbols add AAPL,GOOGL,MSFT
jonbu-ohlcv cli symbols list

# Database operations
jonbu-ohlcv cli migrate up
jonbu-ohlcv cli migrate down
jonbu-ohlcv cli migrate status
```

---

## Phase 2: Real-time Streaming Implementation

### 2.1 WebSocket Server (Week 5)

#### Streaming Infrastructure
```go
// REQ-017 to REQ-020
internal/stream/
├── server.go        # WebSocket server setup
├── client.go        # Client connection management
├── hub.go           # Client hub for message distribution
└── subscription.go  # Symbol/timeframe subscriptions
```

**Streaming Features:**
- **REQ-017**: WebSocket endpoints for real-time data
- **REQ-018**: Symbol and timeframe subscription management
- **REQ-019**: Connection lifecycle handling
- **REQ-020**: Backpressure and flow control

### 2.2 Event-Driven Pipeline (Week 5-6)

#### Worker Architecture
```go
// REQ-031 to REQ-035
internal/worker/
├── pool.go          # Worker pool management
├── symbol_worker.go # Per-symbol processing
├── coordinator.go   # Cross-worker coordination
└── metrics.go       # Performance monitoring
```

**Performance Targets:**
- **REQ-031**: 10k+ events/second per worker
- **REQ-032**: Sub-millisecond aggregation latency
- **REQ-034**: Buffered channels preventing blocking
- **REQ-035**: Memory usage monitoring and cleanup

### 2.3 Channel Architecture (Week 6)

#### Event Flow Design
```go
// Channel-based communication
type Pipeline struct {
    RawEvents     chan MarketEvent     // From data sources
    FilteredEvents chan MarketEvent    // After validation
    Candles       chan Candle         // Aggregated candles
    ClientEvents  chan ClientMessage  // WebSocket distribution
}
```

**Pipeline Stages:**
1. **Ingestion**: Alpaca WebSocket → RawEvents
2. **Validation**: Data validation → FilteredEvents
3. **Aggregation**: Per-symbol workers → Candles
4. **Distribution**: WebSocket hub → Clients

---

## Phase 3: AI Integration Implementation

### 3.1 Technical Indicators Engine (Week 7-8)

#### Indicator Library
```go
// REQ-206 to REQ-210
internal/indicators/
├── trend.go         # SMA, EMA, MACD
├── momentum.go      # RSI, Stochastic, Williams %R
├── volatility.go    # Bollinger Bands, ATR
├── volume.go        # Volume MA, VWAP, OBV
└── cache.go         # Indicator caching system
```

**Indicator Categories:**
- **REQ-206**: Trend indicators (SMA, EMA, MACD)
- **REQ-207**: Momentum indicators (RSI, Stochastic)
- **REQ-208**: Volatility indicators (Bollinger Bands, ATR)
- **REQ-209**: Volume indicators (VWAP, OBV)
- **REQ-210**: Performance-optimized caching

### 3.2 Market Context Analysis (Week 9-10)

#### Pattern Recognition
```go
// REQ-211 to REQ-215
internal/analysis/
├── candlestick.go   # Candlestick pattern detection
├── chart.go         # Chart pattern recognition
├── regime.go        # Market regime identification
├── support.go       # Support/resistance levels
└── trend.go         # Trend strength analysis
```

**Analysis Features:**
- **REQ-211**: Candlestick patterns (doji, hammer, etc.)
- **REQ-212**: Chart patterns (breakouts, reversals)
- **REQ-213**: Market regime detection
- **REQ-214**: Dynamic support/resistance
- **REQ-215**: Trend strength assessment

### 3.3 Enriched Candle Pipeline (Week 11) - On-Demand Strategy

#### AI-Ready Data Structure (Calculated On-Demand)
```go
// REQ-200 to REQ-205: Enriched candles calculated in real-time, NOT stored
type EnrichedCandle struct {
    // Basic OHLCV (from database)
    OHLCV           Candle              `json:"ohlcv"`
    
    // Technical Indicators (calculated on-demand)
    Indicators      IndicatorSet        `json:"indicators"`
    
    // Market Context (calculated on-demand)
    Context         MarketContext       `json:"context"`
    
    // ML Features (calculated on-demand)
    Features        FeatureVector       `json:"features"`
    
    // Enrichment Metadata (processing info)
    ProcessedAt     time.Time           `json:"processed_at"`
    Confidence      float64             `json:"confidence"`
    Metadata        map[string]string   `json:"metadata"`
}
```

**On-Demand Enrichment Pipeline:**
1. **Fetch Raw OHLCV** → From PostgreSQL database (lightweight)
2. **Calculate Indicators** → Real-time computation (0.167ms per candle)
3. **Generate Context** → Market analysis on-demand
4. **Create Features** → Feature vector generation
5. **Return Enriched** → To client/backtesting engine

**Storage Strategy Rationale:**
- **Performance**: 0.167ms enrichment latency makes real-time viable
- **Storage**: Avoid 50x database size increase from storing enriched data
- **Flexibility**: Easy indicator modifications without schema changes
- **Accuracy**: Always use latest enrichment logic for consistency

---

## Phase 4: RAG Integration Implementation

### 4.1 Vector Preparation & Context Generation (Week 12) - On-Demand Processing

#### RAG Data Pipeline (On-Demand Export)
```go
// REQ-221 to REQ-225: On-demand processing, not persistent storage
internal/rag/
├── processor.go     # On-demand enrichment for export
├── context.go       # Context text generation from enriched data
├── metadata.go      # Searchable metadata generation
├── normalizer.go    # Feature normalization during export
└── exporter.go      # Batch export from raw OHLCV + enrichment
```

**RAG Features (On-Demand):**
- **REQ-221**: Vector export preparation (enriched on-demand from raw OHLCV)
- **REQ-222**: Feature normalization (0-1 range) during processing
- **REQ-223**: Context text generation from real-time enrichment
- **REQ-224**: Searchable metadata filtering during export
- **REQ-225**: Batch export for model training

### 4.2 Human-Readable Descriptions (Week 13)

#### Market Narrative Generation
```go
// REQ-226 to REQ-230
internal/narrative/
├── generator.go     # Market description generator
├── templates.go     # Description templates
├── formatter.go     # Output formatting
└── validator.go     # Content validation
```

**Narrative Features:**
- **REQ-226**: Real-time market descriptions
- **REQ-227**: Technical analysis summaries
- **REQ-228**: Pattern recognition explanations
- **REQ-229**: Context-aware narratives
- **REQ-230**: Multi-format output (JSON, text, markdown)

Example on-demand enriched candle for RAG export:
```go
// RAG export function - processes raw OHLCV on-demand
func (r *RAGExporter) PrepareForExport(ohlcv *models.OHLCV, history []*models.OHLCV) (*RAGReadyCandle, error) {
    // 1. Enrich candle on-demand (0.167ms)
    enriched, err := r.enricher.EnrichCandle(ctx, ohlcv, history, nil)
    if err != nil {
        return nil, fmt.Errorf("enrichment failed: %w", err)
    }
    
    // 2. Generate RAG-specific fields
    return &RAGReadyCandle{
        EnrichedCandle:     *enriched,
        Description:        r.generateDescription(enriched),
        EmbeddingVector:    r.generateEmbedding(enriched),
        SearchKeywords:     r.extractKeywords(enriched),
        TechnicalSummary:   r.generateSummary(enriched),
        PatternExplanation: r.explainPatterns(enriched),
    }, nil
}

type RAGReadyCandle struct {
    // Core enriched candle (calculated on-demand)
    EnrichedCandle
    
    // RAG-specific fields (generated for export)
    Description     string              `json:"description"`
    EmbeddingVector []float64           `json:"embedding_vector"`
    SearchKeywords  []string            `json:"search_keywords"`
    TechnicalSummary string             `json:"technical_summary"`
    PatternExplanation string           `json:"pattern_explanation"`
}
```

Example market description:
```
AAPL 1-minute candle at 2025-07-12 14:30:00 EST:
Price: $150.25 (+0.5%), Volume: 1.2M shares
Technical: RSI(70) overbought, MACD bullish crossover
Pattern: Hammer formation suggesting reversal
Trend: Strong uptrend (20-period slope +15°)
Context: Breaking resistance at $150, high volume confirmation
Confidence: 85% based on technical indicators alignment
```

### 4.3 Vector Database Integration (Week 14-15) - Export Pipeline

#### On-Demand Export Pipeline
```go
// Vector database export (not persistent storage)
internal/vectordb/
├── exporter.go      # On-demand vector export from raw OHLCV
├── processor.go     # Batch enrichment for export
├── formatter.go     # Vector formatting and normalization
└── validator.go     # Export validation
```

**Export Features (On-Demand Processing):**
- Fetch raw OHLCV from PostgreSQL database
- Calculate enriched candles using enrichment engine (0.167ms/candle)
- Generate multi-dimensional feature vectors on-demand
- Export to vector database for ML training/inference
- Batch processing for historical data export
- Real-time export for streaming applications

**Benefits of On-Demand Approach:**
- **Storage Efficiency**: No persistent vector storage in primary database
- **Freshness**: Always export with latest indicator calculations
- **Flexibility**: Easy to modify export formats and features
- **Performance**: Sub-millisecond enrichment makes real-time export viable

---

## Phase 5: Production Hardening Implementation

### 5.1 Monitoring & Health Checks (Week 16)

#### Observability Stack
```go
// REQ-081 to REQ-085
internal/monitoring/
├── health.go        # Health check endpoints
├── metrics.go       # Prometheus metrics
├── profiling.go     # Performance profiling
└── alerting.go      # Alert conditions
```

**Monitoring Coverage:**
- **REQ-081**: Component health endpoints
- **REQ-082**: Prometheus metrics export
- **REQ-083**: API response time tracking
- **REQ-084**: Resource usage monitoring
- **REQ-085**: Critical failure alerting

### 5.2 Error Handling & Recovery (Week 16-17)

#### Resilience Patterns
```go
// REQ-091 to REQ-100
internal/resilience/
├── circuit_breaker.go  # Circuit breaker pattern
├── retry.go           # Exponential backoff
├── recovery.go        # Panic recovery
└── degradation.go     # Graceful degradation
```

**Error Categories:**
- **REQ-091**: Transient vs permanent error classification
- **REQ-092**: Exponential backoff for retries
- **REQ-096**: Panic recovery without crashes
- **REQ-097**: Automatic reconnection strategies

### 5.3 Security Implementation (Week 17)

#### Security Measures
```go
// REQ-041 to REQ-045
internal/security/
├── validator.go     # Input validation
├── auth.go         # Authentication middleware
├── ratelimiter.go  # API rate limiting
└── secrets.go      # Secret management
```

**Security Features:**
- **REQ-041**: Comprehensive input validation
- **REQ-042**: SQL injection prevention
- **REQ-043**: API rate limiting
- **REQ-044**: Environment-based secret management
- **REQ-045**: HTTPS enforcement

### 5.4 Deployment & Operations (Week 18)

#### Production Readiness
```go
// REQ-086 to REQ-090
docker/
├── Dockerfile
├── docker-compose.yml
└── k8s/
    ├── deployment.yaml
    ├── service.yaml
    └── configmap.yaml
```

**Operational Features:**
- **REQ-086**: Docker containerization
- **REQ-087**: Graceful shutdown handling
- **REQ-088**: Rolling deployment support
- **REQ-089**: Configuration hot-reloading
- **REQ-090**: Horizontal scaling capabilities

---

## 🎯 Success Metrics & Validation

### Phase 1 Completion Criteria
- [ ] Successfully ingest historical data from Alpaca
- [ ] Store and retrieve OHLCV data via REST API
- [ ] CLI tools for basic operations
- [ ] Database migrations working
- [ ] >80% test coverage for core components

### Phase 2 Completion Criteria
- [ ] Real-time WebSocket streaming operational
- [ ] Handle 10k+ events/second per worker
- [ ] Sub-millisecond aggregation latency
- [ ] Support 100+ concurrent WebSocket clients
- [ ] Multi-timeframe aggregation working

### Phase 3 Completion Criteria
- [ ] 15+ technical indicators calculated in real-time
- [ ] Market context analysis operational
- [ ] Enriched candles generated with <1ms overhead (on-demand, not stored)
- [ ] Pattern recognition working accurately
- [ ] Support 100+ symbols with on-demand enrichment

### Phase 4 Completion Criteria
- [ ] Vector export pipeline operational (on-demand processing)
- [ ] Human-readable market descriptions generated
- [ ] Context text generation operational
- [ ] Batch export functionality for ML training
- [ ] On-demand enrichment integration complete

### Phase 5 Completion Criteria
- [ ] Health checks and monitoring operational
- [ ] Graceful error handling and recovery
- [ ] Security measures implemented
- [ ] Production deployment pipeline
- [ ] 99.9% uptime in staging environment

---

## 🔧 Development Workflow

### Daily Development Process
1. **Requirement Review**: Start each task with specific REQ-XXX references
2. **TDD Approach**: Write tests first, then implementation
3. **Code Generation**: Use GitHub Copilot with requirement templates
4. **Validation**: Ensure all requirements are addressed
5. **Documentation**: Update docs and examples

### Quality Gates
- **Code Review**: All code must be reviewed before merge
- **Test Coverage**: Maintain >80% coverage for business logic
- **Performance Testing**: Benchmark critical paths
- **Security Scanning**: Automated security vulnerability checks
- **Documentation**: Update API docs and architectural decisions

### Risk Mitigation
- **Incremental Delivery**: Each phase builds on previous success
- **Fallback Plans**: Graceful degradation for non-critical features
- **Performance Monitoring**: Early detection of performance regressions
- **External Dependencies**: Circuit breakers for API failures
- **Data Integrity**: Comprehensive validation at all boundaries

## 📊 Architectural Decision: On-Demand Enrichment Strategy

### Decision Summary
**Storage Strategy**: Store only raw OHLCV data in PostgreSQL. Calculate enriched candles on-demand.

### Rationale
- **Performance**: 0.167ms enrichment latency makes real-time calculation viable for all use cases
- **Storage Efficiency**: Avoids 50x database size increase from storing pre-computed enriched data
- **Flexibility**: Easy to modify indicators and analysis without database schema changes
- **Cost Effectiveness**: Minimal compute overhead vs massive storage and I/O costs
- **Data Freshness**: Always uses latest enrichment logic for consistent results

### Implementation Impact
- **Backtesting**: Fetch raw OHLCV + enrich on-demand during analysis
- **Real-time Streaming**: Continue current pattern of live enrichment
- **RAG Export**: Process raw OHLCV through enrichment pipeline for vector database export
- **Database Schema**: Simplified to core OHLCV tables only
- **API Performance**: Sub-millisecond enrichment maintains excellent response times

### Trade-offs Considered
- ✅ **Compute vs Storage**: Chose compute (0.167ms) over storage (50x increase)
- ✅ **Flexibility vs Performance**: Gained flexibility without sacrificing performance
- ✅ **Complexity vs Maintenance**: Reduced database complexity, simplified maintenance
- ✅ **Memory vs Disk**: Better memory utilization, reduced disk I/O

---

## 📚 Dependencies & Tools

### Core Dependencies
```go
// Go 1.21+ (REQ-066)
github.com/rs/zerolog v1.34.0          // Logging (REQ-067)
github.com/gorilla/mux v1.8.1          // HTTP routing (REQ-068)
github.com/spf13/cobra v1.8.0          // CLI (REQ-069)
github.com/spf13/viper v1.17.0         // Config (REQ-070)
github.com/lib/pq v1.10.9              // PostgreSQL
github.com/gorilla/websocket           // WebSocket support
```

### Development Tools
```bash
# Testing and quality
go test -v -race -coverprofile=coverage.out ./...
golangci-lint run
goimports -w .

# Performance profiling
go tool pprof cpu.prof
go tool pprof mem.prof

# Database tools
migrate -path migrations -database postgres://... up
```

### Monitoring Stack
- **Metrics**: Prometheus + Grafana
- **Logging**: ELK Stack or similar
- **Tracing**: OpenTelemetry
- **Health Checks**: Custom endpoints + monitoring

---

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 12+
- Docker & Docker Compose
- Git

### Phase 1 Quick Start
```bash
# 1. Setup project structure
mkdir -p internal/{config,models,database,logger}
mkdir -p internal/fetcher/alpaca
mkdir -p pkg/api/{handlers,middleware,types}
mkdir -p cmd/{cli,server}

# 2. Initialize Go modules
go mod init github.com/ridopark/jonbu-ohlcv
go get github.com/rs/zerolog@v1.34.0
go get github.com/gorilla/mux@v1.8.1
go get github.com/spf13/cobra@v1.8.0

# 3. Setup database
createdb jonbu_ohlcv
migrate -path migrations -database postgres://localhost/jonbu_ohlcv up

# 4. Start development
go run cmd/server/main.go
```

This implementation plan provides a comprehensive roadmap for building a production-ready OHLCV streaming service with advanced AI integration capabilities, ensuring all 230+ requirements are systematically addressed across the four development phases.
