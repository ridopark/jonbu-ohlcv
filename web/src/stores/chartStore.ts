import { create } from 'zustand';
import { persist } from 'zustand/middleware';

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

export interface SupportResistanceLevel {
  price: number;
  type: 'support' | 'resistance';
  strength: number; // 0-100
  touches: number;
  last_touch: number;
  confidence: number; // 0-100
}

export interface CurrentLevelInfo {
  price: number;
  nearest_support?: SupportResistanceLevel;
  nearest_resistance?: SupportResistanceLevel;
  distance_to_support: number;
  distance_to_resistance: number;
  position: 'near_support' | 'near_resistance' | 'middle';
}

export interface SupportResistanceLevels {
  support: SupportResistanceLevel[];
  resistance: SupportResistanceLevel[];
  current?: CurrentLevelInfo;
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
  support_resistance?: SupportResistanceLevels;
}

export type Timeframe = '1m' | '5m' | '15m' | '30m' | '1h' | '4h' | '1d';

interface ChartStore {
  // Current symbol and timeframe
  symbol: string;
  timeframe: Timeframe;
  
  // Chart data
  candles: OHLCVCandle[];
  enrichedCandles: EnrichedCandle[];
  loading: boolean;
  error: string | null;
  
  // Support/Resistance levels
  supportResistanceLevels: SupportResistanceLevels | null;
  showSupportResistance: boolean;
  
  // Chart settings
  showIndicators: boolean;
  indicators: string[];
  chartType: 'candlestick' | 'line' | 'area';
  
  // Real-time data
  isStreaming: boolean;
  lastUpdate: string | null;
  
  // Actions
  setSymbol: (symbol: string) => void;
  setTimeframe: (timeframe: Timeframe) => void;
  setCandles: (candles: OHLCVCandle[]) => void;
  setEnrichedCandles: (candles: EnrichedCandle[]) => void;
  addCandle: (candle: OHLCVCandle) => void;
  addEnrichedCandle: (candle: EnrichedCandle) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setSupportResistanceLevels: (levels: SupportResistanceLevels | null) => void;
  toggleSupportResistance: () => void;
  toggleIndicator: (indicator: string) => void;
  setChartType: (type: 'candlestick' | 'line' | 'area') => void;
  setStreaming: (streaming: boolean) => void;
  clearData: () => void;
}

const DEFAULT_INDICATORS = ['SMA20', 'EMA50', 'RSI', 'Volume'];

export const useChartStore = create<ChartStore>()(
  persist(
    (set, get) => ({
      // Initial state
      symbol: 'AAPL',
      timeframe: '1m',
      candles: [],
      enrichedCandles: [],
      loading: false,
      error: null,
      supportResistanceLevels: null,
      showSupportResistance: true,
      showIndicators: true,
      indicators: DEFAULT_INDICATORS,
      chartType: 'candlestick',
      isStreaming: false,
      lastUpdate: null,
      
      // Actions
      setSymbol: (symbol: string) => {
        set({ symbol, candles: [], enrichedCandles: [], error: null });
      },
      
      setTimeframe: (timeframe: Timeframe) => {
        set({ timeframe, candles: [], enrichedCandles: [], error: null });
      },
      
      setCandles: (candles: OHLCVCandle[]) => {
        set({ 
          candles: candles.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()),
          lastUpdate: new Date().toISOString()
        });
      },
      
      setEnrichedCandles: (enrichedCandles: EnrichedCandle[]) => {
        set({ 
          enrichedCandles: enrichedCandles.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()),
          lastUpdate: new Date().toISOString()
        });
      },
      
      addCandle: (candle: OHLCVCandle) => {
        console.log('🏪 STORE: Adding candle to chart store:', {
          symbol: candle.symbol,
          timestamp: candle.timestamp,
          interval: candle.interval,
          ohlcv: {
            open: candle.open,
            high: candle.high,
            low: candle.low,
            close: candle.close,
            volume: candle.volume
          }
        });
        
        // Validate candle data before adding to store
        if (!candle || 
            typeof candle.open !== 'number' || !isFinite(candle.open) ||
            typeof candle.high !== 'number' || !isFinite(candle.high) ||
            typeof candle.low !== 'number' || !isFinite(candle.low) ||
            typeof candle.close !== 'number' || !isFinite(candle.close) ||
            typeof candle.volume !== 'number' || !isFinite(candle.volume) ||
            candle.open <= 0 || candle.high <= 0 || candle.low <= 0 || candle.close <= 0 ||
            candle.volume < 0) {
          console.warn('🚫 Invalid candle data rejected:', candle);
          return;
        }
        
        const { candles } = get();
        const updatedCandles = [...candles];
        
        // Check if candle already exists (update) or is new (append)
        const existingIndex = updatedCandles.findIndex(
          c => c.timestamp === candle.timestamp && c.symbol === candle.symbol
        );
        
        if (existingIndex >= 0) {
          updatedCandles[existingIndex] = candle;
          console.log('🔄 Updated existing candle at index:', existingIndex);
        } else {
          updatedCandles.push(candle);
          console.log('➕ Added new candle, total candles:', updatedCandles.length);
          // Keep only last 1000 candles for performance
          if (updatedCandles.length > 1000) {
            updatedCandles.shift();
          }
        }
        
        const sortedCandles = updatedCandles.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
        console.log('📊 Final candles array length:', sortedCandles.length);
        
        set({ 
          candles: sortedCandles,
          lastUpdate: new Date().toISOString()
        });
      },
      
      addEnrichedCandle: (candle: EnrichedCandle) => {
        const { enrichedCandles } = get();
        const updatedCandles = [...enrichedCandles];
        
        const existingIndex = updatedCandles.findIndex(
          c => c.timestamp === candle.timestamp && c.symbol === candle.symbol
        );
        
        if (existingIndex >= 0) {
          updatedCandles[existingIndex] = candle;
        } else {
          updatedCandles.push(candle);
          if (updatedCandles.length > 1000) {
            updatedCandles.shift();
          }
        }
        
        set({ 
          enrichedCandles: updatedCandles.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()),
          lastUpdate: new Date().toISOString()
        });
      },
      
      setLoading: (loading: boolean) => set({ loading }),
      
      setError: (error: string | null) => set({ error }),
      
      setSupportResistanceLevels: (supportResistanceLevels: SupportResistanceLevels | null) => 
        set({ supportResistanceLevels }),
      
      toggleSupportResistance: () => {
        const { showSupportResistance } = get();
        set({ showSupportResistance: !showSupportResistance });
      },
      
      toggleIndicator: (indicator: string) => {
        const { indicators } = get();
        const updated = indicators.includes(indicator)
          ? indicators.filter(i => i !== indicator)
          : [...indicators, indicator];
        set({ indicators: updated });
      },
      
      setChartType: (chartType: 'candlestick' | 'line' | 'area') => set({ chartType }),
      
      setStreaming: (isStreaming: boolean) => set({ isStreaming }),
      
      clearData: () => set({ 
        candles: [], 
        enrichedCandles: [], 
        error: null, 
        lastUpdate: null 
      }),
    }),
    {
      name: 'chart-storage',
      partialize: (state) => ({
        symbol: state.symbol,
        timeframe: state.timeframe,
        showIndicators: state.showIndicators,
        indicators: state.indicators,
        chartType: state.chartType,
      }),
    }
  )
);
