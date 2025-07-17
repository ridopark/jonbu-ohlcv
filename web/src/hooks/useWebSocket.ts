import { useEffect, useCallback } from 'react';
import { useChartStore } from '../stores/chartStore';
import { wsClient } from '../utils/websocket';
import type { OHLCVCandle } from '../types';

// Map frontend timeframe format to backend format
const mapTimeframeToBackend = (tf: string): string => {
  const mapping: Record<string, string> = {
    '1m': '1min',
    '5m': '5min',
    '15m': '15min',
    '30m': '30min',
    '1h': '1hour',
    '4h': '4hour',
    '1d': '1day',
  };
  return mapping[tf] || tf;
};

export const useWebSocket = () => {
  const { 
    symbol, 
    timeframe, 
    setStreaming, 
    addCandle, 
    addEnrichedCandle,
    setError 
  } = useChartStore();

  // Handle incoming candle data
  const handleCandle = useCallback((message: any) => {
    try {
      console.log('🔍 Raw candle message received:', message);
      
      // Handle different message formats from backend
      let candleData = null;
      if (message.type === 'candle') {
        // Try different possible locations for candle data
        if (message.candle) {
          candleData = message.candle;
        } else if (message.data) {
          candleData = message.data;
        } else if (message.symbol) {
          // Sometimes the candle data is directly in the message
          candleData = message;
        }
      }
      
      if (candleData) {
        const candle: OHLCVCandle = {
          symbol: candleData.symbol,
          timestamp: candleData.timestamp,
          open: candleData.open,
          high: candleData.high,
          low: candleData.low,
          close: candleData.close,
          volume: candleData.volume,
          interval: candleData.interval || candleData.timeframe || timeframe,
        };
        
        // Validate candle data before processing
        const isValidCandle = 
          typeof candle.open === 'number' && isFinite(candle.open) &&
          typeof candle.high === 'number' && isFinite(candle.high) &&
          typeof candle.low === 'number' && isFinite(candle.low) &&
          typeof candle.close === 'number' && isFinite(candle.close) &&
          typeof candle.volume === 'number' && isFinite(candle.volume);
        
        if (!isValidCandle) {
          console.error('🚫 Received invalid candle data:', {
            candle,
            rawData: candleData,
            validation: {
              open: typeof candle.open === 'number' && isFinite(candle.open),
              high: typeof candle.high === 'number' && isFinite(candle.high),
              low: typeof candle.low === 'number' && isFinite(candle.low),
              close: typeof candle.close === 'number' && isFinite(candle.close),
              volume: typeof candle.volume === 'number' && isFinite(candle.volume)
            }
          });
          return;
        }
        
        console.log('🎯 Frontend received candle:', {
          symbol: candle.symbol,
          interval: candle.interval,
          timestamp: candle.timestamp,
          ohlcv: {
            open: candle.open,
            high: candle.high,
            low: candle.low,
            close: candle.close,
            volume: candle.volume
          },
          currentSubscription: {
            symbol: symbol,
            timeframe: timeframe,
            backendTimeframe: mapTimeframeToBackend(timeframe)
          },
          isMatchingSubscription: candle.symbol === symbol && candle.interval === mapTimeframeToBackend(timeframe),
          rawMessageFormat: {
            hasCandle: !!message.candle,
            hasData: !!message.data,
            hasDirectSymbol: !!message.symbol
          }
        });
        
        // Check if candle matches current subscription
        const backendTimeframe = mapTimeframeToBackend(timeframe);
        if (candle.symbol === symbol && candle.interval === backendTimeframe) {
          console.log('✅ Adding candle to chart store (matches current subscription)');
          addCandle(candle);
        } else {
          console.log('⚠️ Candle does not match current subscription, ignoring', {
            received: { symbol: candle.symbol, interval: candle.interval },
            expected: { symbol, timeframe, backendTimeframe }
          });
        }
      } else {
        console.warn('🔥 Invalid candle message format - no candle data found:', {
          message,
          hasCandle: !!message.candle,
          hasData: !!message.data,
          hasSymbol: !!message.symbol,
          messageKeys: Object.keys(message)
        });
      }
    } catch (error) {
      console.error('💥 Error processing candle data:', error, 'Message:', message);
      setError('Failed to process candle data');
    }
  }, [addCandle, setError, timeframe, symbol]);

  // Handle incoming enriched candle data
  const handleEnrichedCandle = useCallback((message: any) => {
    try {
      console.log('🔮 Raw enriched candle message received:', message);
      console.log('🔍 Current subscription state:', { symbol, timeframe, backendTimeframe: mapTimeframeToBackend(timeframe) });
      console.log('🔍 Message structure analysis:', {
        type: message.type,
        symbol: message.symbol,
        timeframe: message.timeframe,
        interval: message.interval,
        hasData: !!message.data,
        dataKeys: message.data ? Object.keys(message.data) : null,
        dataInterval: message.data?.interval
      });
      
      // Extract the enriched data - backend sends nested structure
      let enrichedData = message;
      if (message.type === 'enriched_candle') {
        enrichedData = message.data || message;
      }
      
      if (enrichedData && enrichedData.ohlcv) {
        const ohlcv = enrichedData.ohlcv;
        const indicators = enrichedData.indicators || {};
        const analysis = enrichedData.analysis || {};
        
        // Validate basic OHLCV data
        const isValidCandle = 
          typeof ohlcv.open === 'number' && isFinite(ohlcv.open) &&
          typeof ohlcv.high === 'number' && isFinite(ohlcv.high) &&
          typeof ohlcv.low === 'number' && isFinite(ohlcv.low) &&
          typeof ohlcv.close === 'number' && isFinite(ohlcv.close) &&
          typeof ohlcv.volume === 'number' && isFinite(ohlcv.volume);
        
        if (!isValidCandle) {
          console.error('🚫 Received invalid OHLCV data in enriched candle:', ohlcv);
          return;
        }

        // Flatten the enriched candle to match frontend type expectations
        const flattenedCandle = {
          // Base OHLCV properties
          symbol: ohlcv.symbol,
          timestamp: ohlcv.timestamp,
          open: ohlcv.open,
          high: ohlcv.high,
          low: ohlcv.low,
          close: ohlcv.close,
          volume: ohlcv.volume,
          interval: message.interval || ohlcv.interval || message.timeframe, // Check top-level first
          
          // Technical indicators
          sma20: indicators.sma_20,
          sma50: indicators.sma_50,
          ema20: indicators.ema_20,
          ema50: indicators.ema_50,
          rsi: indicators.rsi,  // Fixed: was rsi_14, should be rsi
          macd: indicators.macd,
          macd_signal: indicators.macd_signal,
          macd_histogram: indicators.macd_histogram,
          bollinger_upper: indicators.bollinger_upper,
          bollinger_middle: indicators.bollinger_middle,
          bollinger_lower: indicators.bollinger_lower,
          volume_sma: indicators.volume_sma,
          
          // Analysis data
          price_change: analysis.price_change,
          price_change_percent: analysis.price_change_percent,
          is_green: analysis.is_green,
          body_size: analysis.body_size,
          upper_shadow: analysis.upper_shadow,
          lower_shadow: analysis.lower_shadow
        };

        console.log('🎯 Frontend received enriched candle:', {
          symbol: flattenedCandle.symbol,
          interval: flattenedCandle.interval,
          timestamp: flattenedCandle.timestamp,
          indicators: Object.keys(indicators),
          analysis: Object.keys(analysis),
          enrichedData: {
            sma20: flattenedCandle.sma20,
            rsi: flattenedCandle.rsi,
            priceChange: flattenedCandle.price_change
          }
        });
        
        // Check if candle matches current subscription
        const backendTimeframe = mapTimeframeToBackend(timeframe);
        
        console.log('🔍 SUBSCRIPTION MATCH DEBUG:', {
          flattenedCandle: {
            symbol: flattenedCandle.symbol,
            interval: flattenedCandle.interval,
            intervalType: typeof flattenedCandle.interval
          },
          expected: {
            symbol: symbol,
            backendTimeframe: backendTimeframe,
            backendTimeframeType: typeof backendTimeframe
          },
          symbolMatch: flattenedCandle.symbol === symbol,
          intervalMatch: flattenedCandle.interval === backendTimeframe,
          overallMatch: flattenedCandle.symbol === symbol && flattenedCandle.interval === backendTimeframe
        });
        
        if (flattenedCandle.symbol === symbol && flattenedCandle.interval === backendTimeframe) {
          console.log('✅ Adding enriched candle to chart store (matches current subscription)');
          
          // Add to enriched candles store
          addEnrichedCandle(flattenedCandle);
          
          // Also add the base OHLCV data to regular candles store for chart display
          const baseCandle = {
            symbol: flattenedCandle.symbol,
            timestamp: flattenedCandle.timestamp,
            open: flattenedCandle.open,
            high: flattenedCandle.high,
            low: flattenedCandle.low,
            close: flattenedCandle.close,
            volume: flattenedCandle.volume,
            interval: flattenedCandle.interval
          };
          addCandle(baseCandle);
          
          console.log('📊 Added both enriched and base candle to stores');
        } else {
          console.log('⚠️ Enriched candle does not match current subscription, ignoring', {
            received: { symbol: flattenedCandle.symbol, interval: flattenedCandle.interval },
            expected: { symbol, timeframe, backendTimeframe }
          });
        }
      } else {
        console.error('� Received invalid enriched candle structure:', enrichedData);
      }
    } catch (error) {
      console.error('💥 Error processing enriched candle data:', error, 'Message:', message);
      setError('Failed to process enriched candle data');
    }
  }, [addEnrichedCandle, addCandle, setError, timeframe, symbol]);

  // Handle connection status
  const handleConnection = useCallback((status: any) => {
    console.log('🔌 WebSocket connection status:', {
      status: status.status,
      timestamp: new Date().toISOString(),
      currentSubscription: { symbol, timeframe }
    });
    setStreaming(status.status === 'connected');
    
    if (status.status === 'connected') {
      setError(null);
      // Subscribe to current symbol/timeframe when connected
      const backendTimeframe = mapTimeframeToBackend(timeframe);
      console.log('🚀 Auto-subscribing on connection:', { symbol, frontend: timeframe, backend: backendTimeframe });
      wsClient.subscribeToSymbol(symbol, backendTimeframe);
    } else if (status.status === 'error' || status.status === 'failed') {
      console.error('💔 WebSocket connection failed:', status);
      setError('WebSocket connection failed');
      setStreaming(false);
    }
  }, [symbol, timeframe, setStreaming, setError]);

  // Handle errors
  const handleError = useCallback((error: any) => {
    console.error('WebSocket error:', error);
    setError(error.message || 'WebSocket error occurred');
  }, [setError]);

  // Subscribe to WebSocket events
  useEffect(() => {
    const unsubscribeCandle = wsClient.subscribe('candle', handleCandle);
    const unsubscribeEnrichedCandle = wsClient.subscribe('enriched_candle', handleEnrichedCandle);
    const unsubscribeConnection = wsClient.subscribe('connection', handleConnection);
    const unsubscribeError = wsClient.subscribe('error', handleError);

    return () => {
      unsubscribeCandle();
      unsubscribeEnrichedCandle();
      unsubscribeConnection();
      unsubscribeError();
    };
  }, [handleCandle, handleEnrichedCandle, handleConnection, handleError]);

  // Handle symbol/timeframe changes
  useEffect(() => {
    if (wsClient.isConnected) {
      const backendTimeframe = mapTimeframeToBackend(timeframe);
      console.log('🔄 Switching subscription:', {
        from: 'previous subscription',
        to: { symbol, frontend: timeframe, backend: backendTimeframe },
        timestamp: new Date().toISOString()
      });
      wsClient.subscribeToSymbol(symbol, backendTimeframe);
    } else {
      console.log('⏳ WebSocket not connected, deferring subscription for:', { symbol, timeframe });
    }
  }, [symbol, timeframe]);

  // Subscription management functions
  const subscribe = useCallback((sym: string, tf: string) => {
    const backendTimeframe = mapTimeframeToBackend(tf);
    wsClient.subscribeToSymbol(sym, backendTimeframe);
  }, []);

  const unsubscribe = useCallback((sym: string, tf: string) => {
    const backendTimeframe = mapTimeframeToBackend(tf);
    wsClient.unsubscribeFromSymbol(sym, backendTimeframe);
  }, []);

  return {
    isConnected: wsClient.isConnected,
    subscribe,
    unsubscribe,
  };
};

export default useWebSocket;
