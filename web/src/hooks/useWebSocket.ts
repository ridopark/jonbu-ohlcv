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
    const unsubscribeConnection = wsClient.subscribe('connection', handleConnection);
    const unsubscribeError = wsClient.subscribe('error', handleError);

    return () => {
      unsubscribeCandle();
      unsubscribeConnection();
      unsubscribeError();
    };
  }, [handleCandle, handleConnection, handleError]);

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
