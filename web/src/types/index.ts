export interface OHLCVCandle {
  id?: number;
  symbol: string;
  timestamp: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  interval: string;
  created_at?: string;
  updated_at?: string;
}

export interface EnrichedCandle extends OHLCVCandle {
  sma20?: number;
  sma50?: number;
  ema20?: number;
  ema50?: number;
  rsi?: number;
  macd?: number;
  macd_signal?: number;
  macd_histogram?: number;
  bollinger_upper?: number;
  bollinger_middle?: number;
  bollinger_lower?: number;
  volume_sma?: number;
  price_change?: number;
  price_change_percent?: number;
  is_green?: boolean;
  body_size?: number;
  upper_shadow?: number;
  lower_shadow?: number;
}

export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface SymbolInfo {
  symbol: string;
  name: string;
  exchange: string;
  sector?: string;
  industry?: string;
  market_cap?: number;
  last_price?: number;
  change?: number;
  change_percent?: number;
  volume?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ServerStatus {
  status: 'online' | 'offline' | 'error';
  uptime: number;
  memory_usage: number;
  cpu_usage: number;
  active_connections: number;
  last_heartbeat: string;
}

export interface WebSocketMessage {
  type: 'candle' | 'enriched' | 'status' | 'error';
  data: any;
  timestamp: string;
}

export type Timeframe = '1m' | '5m' | '15m' | '30m' | '1h' | '4h' | '1d';

export type ChartType = 'candlestick' | 'line' | 'area';

export type Theme = 'light' | 'dark' | 'system';
