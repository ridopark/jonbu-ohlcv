-- Create symbols table for symbol management
-- REQ-023: Symbol management commands
-- REQ-080: Symbol format validation

CREATE TABLE IF NOT EXISTS symbols (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL UNIQUE CHECK (symbol ~ '^[A-Z]+$'),
    name VARCHAR(255),
    exchange VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index for active symbol lookups
CREATE INDEX IF NOT EXISTS idx_symbols_active ON symbols (symbol) WHERE is_active = true;

-- Create index for exchange-based queries
CREATE INDEX IF NOT EXISTS idx_symbols_exchange ON symbols (exchange);

-- Add comment to table
COMMENT ON TABLE symbols IS 'Tracked stock symbols for OHLCV data collection';
COMMENT ON COLUMN symbols.symbol IS 'Stock ticker symbol (uppercase letters only)';
COMMENT ON COLUMN symbols.name IS 'Full company name';
COMMENT ON COLUMN symbols.exchange IS 'Stock exchange (NYSE, NASDAQ, etc.)';
COMMENT ON COLUMN symbols.is_active IS 'Whether the symbol is actively tracked';
