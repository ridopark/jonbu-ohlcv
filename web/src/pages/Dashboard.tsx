import React from 'react';
import { useChartStore } from '../stores/chartStore';
import useWebSocket from '../hooks/useWebSocket';
import CandlestickChart from '../components/charts/CandlestickChart';
import SymbolSelector from '../components/forms/SymbolSelector';
import TimeframeSelector from '../components/forms/TimeframeSelector';
import MarketStatus from '../components/ui/MarketStatus';
import StatsCards from '../components/ui/StatsCards';

const Dashboard: React.FC = () => {
  const { 
    symbol, 
    timeframe, 
    candles, 
    enrichedCandles, 
    loading, 
    error, 
    isStreaming,
    lastUpdate 
  } = useChartStore();

  // Initialize WebSocket connection
  const { isConnected } = useWebSocket();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-card-foreground">Dashboard</h1>
          <p className="text-muted-foreground">
            Real-time OHLCV data streaming and analysis
          </p>
        </div>
        
        <div className="flex items-center gap-4">
          <MarketStatus />
          
          {/* WebSocket Connection Status */}
          <div className={`flex items-center gap-2 px-3 py-1 rounded-full ${
            isConnected 
              ? 'bg-blue-100 dark:bg-blue-900/20' 
              : 'bg-gray-100 dark:bg-gray-900/20'
          }`}>
            <div className={`w-2 h-2 rounded-full ${
              isConnected ? 'bg-blue-500 animate-pulse' : 'bg-gray-400'
            }`} />
            <span className={`text-sm font-medium ${
              isConnected 
                ? 'text-blue-700 dark:text-blue-400' 
                : 'text-gray-600 dark:text-gray-400'
            }`}>
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
          
          {isStreaming && (
            <div className="flex items-center gap-2 px-3 py-1 bg-green-100 dark:bg-green-900/20 rounded-full">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span className="text-sm text-green-700 dark:text-green-400 font-medium">
                Live
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Controls */}
      <div className="flex flex-col sm:flex-row gap-4 p-4 bg-card rounded-lg border border-border">
        <div className="flex-1">
          <SymbolSelector />
        </div>
        <div className="flex-1">
          <TimeframeSelector />
        </div>
        <div className="flex items-end">
          <button
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
            onClick={() => window.location.reload()}
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <StatsCards />

      {/* Main Chart */}
      <div className="bg-card rounded-lg border border-border p-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-xl font-semibold text-card-foreground">
              {symbol} - {timeframe}
            </h2>
            {lastUpdate && (
              <p className="text-sm text-muted-foreground">
                Last update: {new Date(lastUpdate).toLocaleTimeString()}
              </p>
            )}
          </div>
          
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">
              {candles.length} candles
            </span>
            {enrichedCandles.length > 0 && (
              <span className="text-sm text-muted-foreground">
                • {enrichedCandles.length} enriched
              </span>
            )}
          </div>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-destructive/10 border border-destructive/20 rounded-md">
            <p className="text-sm text-destructive">{error}</p>
          </div>
        )}

        {loading ? (
          <div className="h-96 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        ) : (
          <CandlestickChart />
        )}
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Latest Candles Stream */}
        <div className="bg-card rounded-lg border border-border p-6">
          <h3 className="text-lg font-semibold text-card-foreground mb-4">
            Latest Candles Stream
            <span className="text-sm text-muted-foreground ml-2">
              (Real-time Log)
            </span>
          </h3>
          
          {candles.length > 0 ? (
            <div className="h-64 overflow-y-auto bg-muted/50 rounded-md p-3 space-y-2 font-mono text-xs">
              {candles.slice(-10).reverse().map((candle, index) => {
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                const isGreen = candle.close > candle.open;
                const priceChange = candle.close - candle.open;
                const priceChangePercent = (priceChange / candle.open) * 100;
                
                return (
                  <div key={`candle-${candle.timestamp}-${index}`} className="text-left border-b border-border/50 pb-2">
                    <div className="flex items-center justify-between mb-1">
                      <div className="text-muted-foreground">
                        [{timestamp}] {symbol}
                      </div>
                      <div className={`flex items-center gap-1 ${isGreen ? 'text-green-600' : 'text-red-600'}`}>
                        <div className={`w-2 h-2 rounded-full ${isGreen ? 'bg-green-500' : 'bg-red-500'}`} />
                        <span className="font-medium">${candle.close.toFixed(2)}</span>
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2 text-card-foreground">
                      <div>O: <span className="text-blue-600">${candle.open.toFixed(2)}</span></div>
                      <div>H: <span className="text-green-600">${candle.high.toFixed(2)}</span></div>
                      <div>L: <span className="text-red-600">${candle.low.toFixed(2)}</span></div>
                      <div>Vol: <span className="text-purple-600">{candle.volume.toLocaleString()}</span></div>
                    </div>
                    
                    <div className="mt-1 text-right">
                      <span className={`text-xs ${isGreen ? 'text-green-600' : 'text-red-600'}`}>
                        {isGreen ? '+' : ''}{priceChange.toFixed(2)} ({priceChangePercent.toFixed(2)}%)
                      </span>
                    </div>
                  </div>
                );
              })}
              
              {/* Auto-scroll indicator */}
              <div className="text-center text-muted-foreground text-xs pt-2">
                ↑ Latest candles above • Streaming live updates
              </div>
            </div>
          ) : (
            <div className="h-64 flex items-center justify-center bg-muted/50 rounded-md">
              <div className="text-center">
                <div className="text-2xl mb-2">📊</div>
                <p className="text-sm text-muted-foreground">
                  Waiting for candle data...
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  OHLCV candles will stream here in real-time
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Technical Indicators Stream */}
        <div className="bg-card rounded-lg border border-border p-6">
          <h3 className="text-lg font-semibold text-card-foreground mb-4">
            Technical Indicators Stream
            <span className="text-sm text-muted-foreground ml-2">
              (Real-time Log)
            </span>
          </h3>
          
          {enrichedCandles.length > 0 ? (
            <div className="h-64 overflow-y-auto bg-muted/50 rounded-md p-3 space-y-2 font-mono text-xs">
              {enrichedCandles.slice(-10).reverse().map((candle, index) => {
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                return (
                  <div key={`indicator-${candle.timestamp}-${index}`} className="text-left border-b border-border/50 pb-2">
                    <div className="text-muted-foreground mb-1">
                      [{timestamp}] {symbol} Indicators:
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2 text-card-foreground">
                      {/* SMA 20 */}
                      {typeof candle.sma20 === 'number' && (
                        <div>SMA20: <span className="text-blue-600">${candle.sma20.toFixed(2)}</span></div>
                      )}
                      
                      {/* SMA 50 */}
                      {typeof candle.sma50 === 'number' && (
                        <div>SMA50: <span className="text-purple-600">${candle.sma50.toFixed(2)}</span></div>
                      )}
                      
                      {/* RSI */}
                      {typeof candle.rsi === 'number' ? (
                        <div>RSI: <span className={
                          candle.rsi > 70 ? 'text-red-600' : 
                          candle.rsi < 30 ? 'text-green-600' : 
                          'text-yellow-600'
                        }>{candle.rsi.toFixed(1)}</span></div>
                      ) : (
                        <div>RSI: <span className="text-gray-500">N/A</span></div>
                      )}
                      
                      {/* MACD */}
                      {candle.macd && typeof candle.macd === 'object' && 'line' in candle.macd && typeof (candle.macd as any).line === 'number' ? (
                        <div>MACD: <span className="text-cyan-600">{((candle.macd as any).line as number).toFixed(4)}</span></div>
                      ) : typeof candle.macd === 'number' ? (
                        <div>MACD: <span className="text-cyan-600">{candle.macd.toFixed(4)}</span></div>
                      ) : (
                        <div>MACD: <span className="text-gray-500">N/A</span></div>
                      )}
                    </div>
                  </div>
                );
              })}
              
              {/* Auto-scroll indicator */}
              <div className="text-center text-muted-foreground text-xs pt-2">
                ↑ Latest indicators above • Streaming live updates
              </div>
            </div>
          ) : (
            <div className="h-64 flex items-center justify-center bg-muted/50 rounded-md">
              <div className="text-center">
                <div className="text-2xl mb-2">📊</div>
                <p className="text-sm text-muted-foreground">
                  Waiting for enriched data...
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  Technical indicators will stream here in real-time
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
