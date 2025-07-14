const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const WS_BASE_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';

export const config = {
  api: {
    baseURL: API_BASE_URL,
    endpoints: {
      candles: '/api/candles',
      enriched: '/api/enriched',
      symbols: '/api/symbols',
      health: '/api/health',
      status: '/api/status',
    },
  },
  websocket: {
    url: `${WS_BASE_URL}/ws`,
    reconnectInterval: 5000,
    maxReconnectAttempts: 10,
  },
  chart: {
    defaultSymbol: 'AAPL',
    defaultTimeframe: '1m' as const,
    maxCandles: 1000,
    updateInterval: 6000, // 6 seconds to match mock service 10x speed
  },
  app: {
    name: 'Jonbu OHLCV',
    version: '1.0.0',
    phase: 'Phase 4 - Web Frontend',
  },
} as const;
