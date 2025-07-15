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
        {/* Latest Candles */}
        <div className="bg-card rounded-lg border border-border p-6">
          <h3 className="text-lg font-semibold text-card-foreground mb-4">
            Latest Candles
          </h3>
          
          <div className="space-y-2">
            {candles.slice(-5).reverse().map((candle, index) => (
              <div 
                key={`${candle.timestamp}-${index}`}
                className="flex items-center justify-between p-3 bg-muted rounded-md"
              >
                <div className="flex items-center gap-3">
                  <div className={`w-3 h-3 rounded-full ${
                    candle.close > candle.open ? 'bg-green-500' : 'bg-red-500'
                  }`} />
                  <div>
                    <p className="text-sm font-medium text-card-foreground">
                      {new Date(candle.timestamp).toLocaleTimeString()}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Vol: {candle.volume.toLocaleString()}
                    </p>
                  </div>
                </div>
                
                <div className="text-right">
                  <p className="text-sm font-medium text-card-foreground">
                    ${candle.close.toFixed(2)}
                  </p>
                  <p className={`text-xs ${
                    candle.close > candle.open ? 'text-green-600' : 'text-red-600'
                  }`}>
                    {candle.close > candle.open ? '+' : ''}
                    {((candle.close - candle.open) / candle.open * 100).toFixed(2)}%
                  </p>
                </div>
              </div>
            ))}
            
            {candles.length === 0 && (
              <p className="text-sm text-muted-foreground text-center py-4">
                No candle data available
              </p>
            )}
          </div>
        </div>

        {/* Technical Indicators Summary */}
        <div className="bg-card rounded-lg border border-border p-6">
          <h3 className="text-lg font-semibold text-card-foreground mb-4">
            Technical Indicators
          </h3>
          
          {enrichedCandles.length > 0 ? (
            <div className="space-y-3">
              {(() => {
                const latest = enrichedCandles[enrichedCandles.length - 1];
                return (
                  <>
                    {latest.sma20 && (
                      <div className="flex justify-between">
                        <span className="text-sm text-muted-foreground">SMA 20</span>
                        <span className="text-sm font-medium text-card-foreground">
                          ${latest.sma20.toFixed(2)}
                        </span>
                      </div>
                    )}
                    
                    {latest.ema50 && (
                      <div className="flex justify-between">
                        <span className="text-sm text-muted-foreground">EMA 50</span>
                        <span className="text-sm font-medium text-card-foreground">
                          ${latest.ema50.toFixed(2)}
                        </span>
                      </div>
                    )}
                    
                    {latest.rsi && (
                      <div className="flex justify-between">
                        <span className="text-sm text-muted-foreground">RSI</span>
                        <span className={`text-sm font-medium ${
                          latest.rsi > 70 ? 'text-red-600' : 
                          latest.rsi < 30 ? 'text-green-600' : 
                          'text-card-foreground'
                        }`}>
                          {latest.rsi.toFixed(1)}
                        </span>
                      </div>
                    )}
                    
                    {latest.macd && (
                      <div className="flex justify-between">
                        <span className="text-sm text-muted-foreground">MACD</span>
                        <span className="text-sm font-medium text-card-foreground">
                          {latest.macd.toFixed(4)}
                        </span>
                      </div>
                    )}
                  </>
                );
              })()}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              No enriched data available
            </p>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
