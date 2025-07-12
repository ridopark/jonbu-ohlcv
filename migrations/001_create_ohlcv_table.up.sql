-- REQ-012: Database migrations for schema management
-- REQ-071: OHLCV structure with all required fields
-- REQ-072: Timestamps in market timezone
-- REQ-073: Prices as decimal/float64
-- REQ-074: Volume as integer/int64

CREATE TABLE IF NOT EXISTS ohlcv (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    open DECIMAL(15,4) NOT NULL CHECK (open > 0),
    high DECIMAL(15,4) NOT NULL CHECK (high > 0),
    low DECIMAL(15,4) NOT NULL CHECK (low > 0),
    close DECIMAL(15,4) NOT NULL CHECK (close > 0),
    volume BIGINT NOT NULL CHECK (volume >= 0),
    timeframe VARCHAR(10) NOT NULL CHECK (timeframe IN ('1m', '5m', '15m', '1h', '4h', '1d')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create unique constraint to prevent duplicate entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_ohlcv_symbol_timestamp_timeframe 
ON ohlcv (symbol, timestamp, timeframe);

-- Create index for efficient symbol-based queries
CREATE INDEX IF NOT EXISTS idx_ohlcv_symbol_timeframe_timestamp 
ON ohlcv (symbol, timeframe, timestamp DESC);

-- Create index for time-range queries
CREATE INDEX IF NOT EXISTS idx_ohlcv_timestamp 
ON ohlcv (timestamp);

-- Add constraint to ensure high >= low
ALTER TABLE ohlcv ADD CONSTRAINT chk_ohlcv_high_low CHECK (high >= low);

-- Add constraint to ensure prices are reasonable (not zero, not negative)
ALTER TABLE ohlcv ADD CONSTRAINT chk_ohlcv_prices_positive 
CHECK (open > 0 AND high > 0 AND low > 0 AND close > 0);

-- Add comment to table
COMMENT ON TABLE ohlcv IS 'OHLCV (Open, High, Low, Close, Volume) market data storage';
COMMENT ON COLUMN ohlcv.symbol IS 'Stock ticker symbol (e.g., AAPL, GOOGL)';
COMMENT ON COLUMN ohlcv.timestamp IS 'Market data timestamp in market timezone';
COMMENT ON COLUMN ohlcv.timeframe IS 'Candle timeframe (1m, 5m, 15m, 1h, 4h, 1d)';
COMMENT ON COLUMN ohlcv.volume IS 'Trading volume for the time period';
