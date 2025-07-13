# GitHub Copilot Requirements for jonbu-ohlcv

## Project Requirements

### 1. Core Functional Requirements

#### Data Ingestion
- **REQ-001**: System MUST ingest real-time OHLCV data from Alpaca API WebSocket
- **REQ-002**: System MUST support multiple data providers (Alpaca, Polygon, etc.)
- **REQ-003**: System MUST handle connection failures with automatic reconnection
- **REQ-004**: System MUST implement rate limiting and backoff strategies
- **REQ-005**: System MUST validate incoming data for integrity and completeness

#### Data Processing
- **REQ-006**: System MUST aggregate tick data into OHLCV candles for multiple timeframes (1m, 5m, 15m, 1h, 1d)
- **REQ-007**: System MUST handle out-of-order events with time-based ordering
- **REQ-008**: System MUST implement incremental aggregation for real-time processing
- **REQ-009**: System MUST support per-symbol worker processes for parallel processing
- **REQ-010**: System MUST emit completed candles via channels

#### Data Storage
- **REQ-011**: System MUST store raw OHLCV data in PostgreSQL database
- **REQ-012**: System MUST implement database migrations for schema management
- **REQ-013**: System MUST use prepared statements for optimal performance
- **REQ-014**: System MUST handle database connection pooling
- **REQ-015**: System MUST implement proper transaction handling
- **STORAGE STRATEGY**: Enriched candles are NOT stored in database - calculated on-demand (0.167ms latency)

#### API & Streaming
- **REQ-016**: System MUST provide REST API endpoints for historical data access
- **REQ-017**: System MUST provide WebSocket endpoints for real-time streaming
- **REQ-018**: System MUST support client subscriptions by symbol and timeframe
- **REQ-019**: System MUST handle WebSocket client connections and disconnections
- **REQ-020**: System MUST implement backpressure handling for streaming

#### CLI Tool
- **REQ-021**: System MUST provide CLI for historical data fetching
- **REQ-022**: System MUST support multiple output formats (JSON, table, CSV)
- **REQ-023**: System MUST provide symbol management commands
- **REQ-024**: System MUST provide database migration commands
- **REQ-025**: System MUST validate user input and provide helpful error messages

### 2. Technical Requirements

#### Architecture
- **REQ-026**: System MUST follow Go clean architecture principles
- **REQ-027**: System MUST use dependency injection for testability
- **REQ-028**: System MUST separate concerns (transport, business logic, data access)
- **REQ-029**: System MUST implement pluggable data provider interfaces
- **REQ-030**: System MUST use event-driven pipeline architecture

#### Performance
- **REQ-031**: System MUST process 10k+ events per second per worker
- **REQ-032**: System MUST maintain sub-millisecond aggregation latency
- **REQ-033**: System MUST handle graceful shutdown within 5 seconds
- **REQ-034**: System MUST use buffered channels to prevent goroutine blocking
- **REQ-035**: System MUST monitor memory usage and prevent leaks

#### Reliability
- **REQ-036**: System MUST implement health checks for all components
- **REQ-037**: System MUST handle panics gracefully with recovery
- **REQ-038**: System MUST provide circuit breakers for external API calls
- **REQ-039**: System MUST implement proper error wrapping and context
- **REQ-040**: System MUST support graceful degradation during failures

#### Security
- **REQ-041**: System MUST validate all user input
- **REQ-042**: System MUST use parameterized queries to prevent SQL injection
- **REQ-043**: System MUST implement rate limiting for API endpoints
- **REQ-044**: System MUST store secrets in environment variables only
- **REQ-045**: System MUST use HTTPS for all external communications

### 3. Code Quality Requirements

#### Logging
- **REQ-046**: All components MUST use zerolog for structured logging
- **REQ-047**: All operations MUST include correlation IDs for tracing
- **REQ-048**: All errors MUST be logged with appropriate context
- **REQ-049**: All performance metrics MUST be logged for monitoring
- **REQ-050**: Log levels MUST be configurable per environment

#### Testing
- **REQ-051**: All business logic MUST have unit tests with >80% coverage
- **REQ-052**: All database operations MUST have integration tests
- **REQ-053**: All API endpoints MUST have integration tests
- **REQ-054**: All tests MUST use table-driven test patterns
- **REQ-055**: All external dependencies MUST be mocked in unit tests

#### Documentation
- **REQ-056**: All exported functions MUST have Go doc comments
- **REQ-057**: All interfaces MUST be documented with usage examples
- **REQ-058**: All configuration options MUST be documented
- **REQ-059**: All API endpoints MUST be documented with OpenAPI spec
- **REQ-060**: All error conditions MUST be documented

### 4. Configuration Requirements

#### Environment Variables
- **REQ-061**: System MUST support .env files for development
- **REQ-062**: System MUST validate configuration on startup
- **REQ-063**: System MUST provide sensible defaults for optional settings
- **REQ-064**: System MUST support multiple environments (dev, staging, prod)
- **REQ-065**: System MUST mask sensitive values in logs

#### Dependencies
- **REQ-066**: System MUST use Go 1.21+ for language features
- **REQ-067**: System MUST use zerolog for high-performance logging
- **REQ-068**: System MUST use gorilla/mux for HTTP routing
- **REQ-069**: System MUST use cobra for CLI framework
- **REQ-070**: System MUST use viper for configuration management

### 5. Data Model Requirements

#### OHLCV Structure
- **REQ-071**: OHLCV MUST include symbol, timestamp, open, high, low, close, volume
- **REQ-072**: Timestamps MUST be in market timezone (America/New_York)
- **REQ-073**: Prices MUST be stored as decimal/float64 with proper precision
- **REQ-074**: Volume MUST be stored as integer/int64
- **REQ-075**: All fields MUST have appropriate JSON and database tags

#### Market Data
- **REQ-076**: System MUST handle market hours (9:30 AM - 4:00 PM EST)
- **REQ-077**: System MUST handle pre-market and after-hours sessions
- **REQ-078**: System MUST handle market holidays and closures
- **REQ-079**: System MUST handle stock splits and dividend adjustments
- **REQ-080**: System MUST validate symbol formats (uppercase letters only)

### 6. Operational Requirements

#### Monitoring
- **REQ-081**: System MUST provide health check endpoints
- **REQ-082**: System MUST expose metrics for Prometheus monitoring
- **REQ-083**: System MUST track API response times and error rates
- **REQ-084**: System MUST monitor goroutine count and memory usage
- **REQ-085**: System MUST alert on critical failures

#### Deployment
- **REQ-086**: System MUST support Docker containerization
- **REQ-087**: System MUST support graceful shutdown signals
- **REQ-088**: System MUST support rolling deployments
- **REQ-089**: System MUST support configuration hot-reloading
- **REQ-090**: System MUST support horizontal scaling

### 7. Error Handling Requirements

#### Error Categories
- **REQ-091**: System MUST distinguish between transient and permanent errors
- **REQ-092**: System MUST implement exponential backoff for retries
- **REQ-093**: System MUST provide meaningful error messages to users
- **REQ-094**: System MUST wrap errors with context for debugging
- **REQ-095**: System MUST handle timeout errors gracefully

#### Recovery
- **REQ-096**: System MUST recover from panics without crashing
- **REQ-097**: System MUST reconnect to data sources automatically
- **REQ-098**: System MUST resume processing from last known state
- **REQ-099**: System MUST notify operators of critical failures
- **REQ-100**: System MUST maintain service availability during failures

### 8. AI & ML Integration Requirements (Phase 4)

#### Enriched Candle Processing
- **REQ-200**: System MUST generate enriched candles with technical indicators in real-time
- **REQ-201**: System MUST calculate 15+ technical indicators with <1ms latency per candle
- **REQ-202**: System MUST analyze market context including trend, volatility, and patterns
- **REQ-203**: System MUST generate human-readable market summaries for RAG systems
- **REQ-204**: System MUST prepare vector embeddings for ML model consumption
- **REQ-205**: System MUST integrate enriched candles into streaming pipeline without affecting basic OHLCV performance

#### Technical Indicators
- **REQ-206**: System MUST support trend indicators (SMA, EMA, MACD)
- **REQ-207**: System MUST support momentum indicators (RSI, Stochastic, Williams %R)
- **REQ-208**: System MUST support volatility indicators (Bollinger Bands, ATR)
- **REQ-209**: System MUST support volume indicators (Volume MA, VWAP, OBV)
- **REQ-210**: System MUST cache indicator calculations for performance optimization

#### Market Context Analysis
- **REQ-211**: System MUST identify candlestick patterns (doji, hammer, shooting star, etc.)
- **REQ-212**: System MUST identify chart patterns (breakouts, reversals, continuations)
- **REQ-213**: System MUST determine market regime (accumulation, markup, distribution, markdown)
- **REQ-214**: System MUST calculate dynamic support and resistance levels
- **REQ-215**: System MUST assess trend strength and direction

#### AI Feature Generation
- **REQ-216**: System MUST generate normalized feature vectors for ML models
- **REQ-217**: System MUST create boolean features for pattern recognition
- **REQ-218**: System MUST generate trading signals based on indicator combinations
- **REQ-219**: System MUST produce contextual text descriptions for RAG retrieval
- **REQ-220**: System MUST include metadata for filtering and categorization

#### Vector Embedding Preparation (On-Demand Processing)
- **REQ-221**: System MUST prepare candles for vector database export on-demand (not stored)
- **REQ-222**: System MUST normalize all numerical features to 0-1 range during processing
- **REQ-223**: System MUST generate comprehensive context text for semantic search on request
- **REQ-224**: System MUST include searchable metadata for filtered retrieval during export
- **REQ-225**: System MUST support batch export for model training from raw OHLCV + enrichment

#### Performance & Scalability
- **REQ-226**: Enriched candle processing MUST NOT impact basic OHLCV latency
- **REQ-227**: System MUST handle enrichment for 100+ symbols simultaneously
- **REQ-228**: System MUST maintain <10MB memory overhead per symbol for indicators
- **REQ-229**: System MUST support configurable indicator history windows
- **REQ-230**: System MUST provide graceful degradation when enrichment fails

---

## Implementation Priorities

### Phase 1: Core Infrastructure (High Priority)
- Data ingestion from Alpaca API
- Basic OHLCV aggregation
- PostgreSQL storage
- REST API endpoints
- CLI tool basics

### Phase 2: Streaming & Real-time (Medium Priority)
- WebSocket streaming
- Per-symbol workers
- Multi-timeframe aggregation
- Client subscription management
- Real-time performance optimization

### Phase 3: Production Features (Medium Priority)
- Comprehensive monitoring
- Error recovery mechanisms
- Performance optimizations
- Security hardening
- Deployment automation

### Phase 4: AI & Trading Intelligence (Advanced Features)
- **Technical Indicator Engine**: Real-time calculation of 15+ indicators
- **Market Context Analysis**: Pattern recognition and regime detection
- **Enriched Candle Streaming**: AI-ready data with contextual information
- **Vector Embedding Preparation**: RAG system integration
- **Trading Signal Generation**: Multi-indicator signal combinations
- **ML Feature Engineering**: Normalized vectors for model consumption
- **Semantic Search Support**: Human-readable market context
- **AI Pipeline Integration**: Non-blocking enrichment processing

---

## Success Criteria

### Performance Metrics
- Process 10,000+ market events per second
- Maintain <1ms aggregation latency
- Support 1,000+ concurrent WebSocket connections
- Achieve 99.9% uptime in production
- Memory usage <512MB per 100 symbols

### Quality Metrics
- >90% test coverage for critical paths
- Zero known security vulnerabilities
- <1% error rate for API endpoints
- <500ms API response time (95th percentile)
- Zero data loss during normal operations

### Operational Metrics
- <5 second startup time
- <5 second graceful shutdown time
- <1 minute deployment time
- <10 minutes recovery from failures
- 24/7 monitoring and alerting coverage

### AI Integration Metrics (Phase 4)
- Generate enriched candles for 100+ symbols in real-time
- Calculate 15+ technical indicators with <1ms overhead per candle
- Maintain 95% accuracy for pattern recognition
- Support 10,000+ enriched candles per minute for vector storage
- Provide <100ms latency for RAG system queries
- Generate contextually accurate market summaries
- Achieve >80% correlation between signals and market moves
