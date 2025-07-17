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
    lastUpdate,
    showSupportResistance,
    toggleSupportResistance
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
        <div className="flex items-end gap-2">
          <button
            className={`px-4 py-2 rounded-md transition-colors ${
              showSupportResistance 
                ? 'bg-green-600 text-white hover:bg-green-700' 
                : 'bg-gray-600 text-white hover:bg-gray-700'
            }`}
            onClick={toggleSupportResistance}
            title="Toggle Support/Resistance Lines"
          >
            S/R
          </button>
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
            <div className="h-64 overflow-y-auto bg-muted/50 rounded-md p-3 space-y-3 font-mono text-xs">
              {enrichedCandles.slice(-8).reverse().map((candle, index) => {
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                return (
                  <div key={`indicator-${candle.timestamp}-${index}`} className="text-left border-b border-border/50 pb-3">
                    <div className="text-muted-foreground mb-2">
                      [{timestamp}] {symbol} Indicators:
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2 text-card-foreground mb-2">
                      {/* Trend Indicators */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        return (
                          <>
                            {typeof indicators?.sma_20 === 'number' && (
                              <div>SMA20: <span className="text-blue-600">${indicators.sma_20.toFixed(2)}</span></div>
                            )}
                            
                            {typeof indicators?.sma_50 === 'number' && (
                              <div>SMA50: <span className="text-purple-600">${indicators.sma_50.toFixed(2)}</span></div>
                            )}
                            
                            {typeof indicators?.ema_12 === 'number' && (
                              <div>EMA12: <span className="text-green-600">${indicators.ema_12.toFixed(2)}</span></div>
                            )}
                            
                            {typeof indicators?.ema_26 === 'number' && (
                              <div>EMA26: <span className="text-indigo-600">${indicators.ema_26.toFixed(2)}</span></div>
                            )}
                          </>
                        );
                      })()}

                      {/* Momentum Indicators */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const rsi = indicators?.rsi;
                        return typeof rsi === 'number' ? (
                          <div>RSI: <span className={
                            rsi > 70 ? 'text-red-600' : 
                            rsi < 30 ? 'text-green-600' : 
                            'text-yellow-600'
                          }>{rsi.toFixed(1)}</span></div>
                        ) : (
                          <div>RSI: <span className="text-gray-500">N/A</span></div>
                        );
                      })()}

                      {/* Williams %R */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const williamsR = indicators?.williams_r;
                        return williamsR && typeof williamsR === 'number' && (
                          <div>Williams%R: <span className={
                            williamsR > -20 ? 'text-red-600' : 
                            williamsR < -80 ? 'text-green-600' : 
                            'text-yellow-600'
                          }>{williamsR.toFixed(1)}</span></div>
                        );
                      })()}

                      {/* MACD */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const macdData = indicators?.macd;
                        const macdLine = macdData?.line;
                        const macdSignal = macdData?.signal;
                        const histogram = macdData?.histogram;
                        
                        // Quick signal determination for log
                        let quickSignal = 'Neutral';
                        if (macdLine !== undefined && macdSignal !== undefined && histogram !== undefined) {
                          if (macdLine > macdSignal && histogram > 0) {
                            quickSignal = 'Bullish';
                          } else if (macdLine < macdSignal && histogram < 0) {
                            quickSignal = 'Bearish';
                          }
                        }
                        
                        return macdData && macdLine !== undefined ? (
                          <div>MACD: <span className="text-cyan-600">{macdLine.toFixed(4)}</span> 
                            <span className={`ml-1 text-xs ${
                              quickSignal === 'Bullish' ? 'text-green-600' : 
                              quickSignal === 'Bearish' ? 'text-red-600' : 
                              'text-gray-500'
                            }`}>
                              ({quickSignal})
                            </span>
                          </div>
                        ) : (
                          <div>MACD: <span className="text-gray-500">N/A</span></div>
                        );
                      })()}

                      {/* Volatility Indicators */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const atr = indicators?.atr;
                        return atr && typeof atr === 'number' && (
                          <div>ATR: <span className="text-orange-600">{atr.toFixed(3)}</span></div>
                        );
                      })()}
                    </div>

                    {/* Advanced Indicators Section */}
                    <div className="mt-2 grid grid-cols-1 gap-1 text-xs">
                      {/* Trend Analysis */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const trendDirection = indicators?.trend_direction;
                        const trendStrength = indicators?.trend_strength;
                        return (trendDirection || trendStrength) && (
                          <div className="flex justify-between items-center">
                            <span className="text-gray-600 dark:text-gray-400">Trend:</span>
                            <div className="flex items-center gap-2">
                              {trendDirection && (
                                <span className={`font-medium ${
                                  trendDirection === 'bullish' ? 'text-green-600' :
                                  trendDirection === 'bearish' ? 'text-red-600' :
                                  'text-gray-600'
                                }`}>
                                  {trendDirection.toUpperCase()}
                                </span>
                              )}
                              {typeof trendStrength === 'number' && (
                                <span className="text-blue-600">
                                  ({trendStrength.toFixed(0)}%)
                                </span>
                              )}
                            </div>
                          </div>
                        );
                      })()}

                      {/* Momentum Analysis */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const momentumDirection = indicators?.momentum_direction;
                        const momentumStrength = indicators?.momentum_strength;
                        return (momentumDirection || momentumStrength) && (
                          <div className="flex justify-between items-center">
                            <span className="text-gray-600 dark:text-gray-400">Momentum:</span>
                            <div className="flex items-center gap-2">
                              {momentumDirection && (
                                <span className={`font-medium ${
                                  momentumDirection === 'bullish' ? 'text-green-600' :
                                  momentumDirection === 'bearish' ? 'text-red-600' :
                                  'text-gray-600'
                                }`}>
                                  {momentumDirection.toUpperCase()}
                                </span>
                              )}
                              {typeof momentumStrength === 'number' && (
                                <span className="text-purple-600">
                                  ({momentumStrength.toFixed(0)}%)
                                </span>
                              )}
                            </div>
                          </div>
                        );
                      })()}

                      {/* Volatility Analysis */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const volatilityLevel = indicators?.volatility_level;
                        const volatilityPercent = indicators?.volatility_percent;
                        return (volatilityLevel || volatilityPercent) && (
                          <div className="flex justify-between items-center">
                            <span className="text-gray-600 dark:text-gray-400">Volatility:</span>
                            <div className="flex items-center gap-2">
                              {volatilityLevel && (
                                <span className={`font-medium ${
                                  volatilityLevel === 'high' ? 'text-red-600' :
                                  volatilityLevel === 'low' ? 'text-green-600' :
                                  'text-yellow-600'
                                }`}>
                                  {volatilityLevel.toUpperCase()}
                                </span>
                              )}
                              {typeof volatilityPercent === 'number' && (
                                <span className="text-orange-600">
                                  ({volatilityPercent.toFixed(1)}%)
                                </span>
                              )}
                            </div>
                          </div>
                        );
                      })()}

                      {/* Bollinger Bands */}
                      {(() => {
                        const indicators = (candle as any).data?.indicators;
                        const bollingerBands = indicators?.bollinger_bands;
                        return bollingerBands?.upper && bollingerBands?.lower && (
                          <div className="flex justify-between items-center">
                            <span className="text-gray-600 dark:text-gray-400">Bollinger:</span>
                            <span className="text-pink-600 text-xs">
                              ${typeof bollingerBands.lower === 'number' ? bollingerBands.lower.toFixed(2) : 'N/A'} - 
                              ${typeof bollingerBands.upper === 'number' ? bollingerBands.upper.toFixed(2) : 'N/A'}
                            </span>
                          </div>
                        );
                      })()}
                    </div>

                    {/* Support & Resistance Levels */}
                    {(() => {
                      const analysis = (candle as any).data?.analysis;
                      const supportResistance = analysis?.support_resistance;
                      return supportResistance && showSupportResistance && (
                        <div className="mt-2 p-2 bg-gray-100 dark:bg-gray-800 rounded border-l-4 border-orange-500">
                          <div className="text-orange-600 font-semibold mb-1">📊 S/R Analysis:</div>
                          
                          {/* Current Position */}
                          {supportResistance.current && (
                            <div className="mb-1">
                              <span className="text-gray-600 dark:text-gray-400">Position: </span>
                              <span className={`font-medium ${
                                supportResistance.current.position === 'near_support' ? 'text-green-600' :
                                supportResistance.current.position === 'near_resistance' ? 'text-red-600' :
                                'text-blue-600'
                            }`}>
                              {supportResistance.current.position?.replace('_', ' ').toUpperCase()}
                            </span>
                          </div>
                        )}

                        {/* Support Levels */}
                        {supportResistance.support && supportResistance.support.length > 0 && (
                          <div className="mb-1">
                            <span className="text-green-600">Support:</span>
                            {supportResistance.support.slice(0, 2).map((level: any, i: number) => (
                              <span key={i} className="ml-1 text-green-700 dark:text-green-400">
                                ${level.price?.toFixed(2)}
                                <span className="text-gray-500 text-xs">
                                  ({level.confidence?.toFixed(0)}%,{level.touches}x)
                                </span>
                                {i < Math.min(supportResistance.support.length, 2) - 1 && ', '}
                              </span>
                            ))}
                          </div>
                        )}

                        {/* Resistance Levels */}
                        {supportResistance.resistance && supportResistance.resistance.length > 0 && (
                          <div>
                            <span className="text-red-600">Resistance:</span>
                            {supportResistance.resistance.slice(0, 2).map((level: any, i: number) => (
                              <span key={i} className="ml-1 text-red-700 dark:text-red-400">
                                ${level.price?.toFixed(2)}
                                <span className="text-gray-500 text-xs">
                                  ({level.confidence?.toFixed(0)}%,{level.touches}x)
                                </span>
                                {i < Math.min(supportResistance.resistance.length, 2) - 1 && ', '}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    );
                    })()}

                    {/* Enhanced Market Analysis Section */}
                    {(() => {
                      const signals = (candle as any).data?.signals;
                      const analysis = (candle as any).data?.analysis;
                      return (signals || analysis) && (
                        <div className="mt-3 p-3 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
                          <div className="text-blue-700 dark:text-blue-300 font-semibold mb-2 text-sm">🎯 Market Analysis</div>
                          
                          {/* Overall Signal */}
                          {signals && (
                            <div className="grid grid-cols-2 gap-2 text-xs mb-2">
                              <div className="flex justify-between">
                                <span className="text-gray-600 dark:text-gray-400">Signal:</span>
                                <span className={`font-medium ${
                                  signals.overall_signal === 'bullish' ? 'text-green-600' :
                                  signals.overall_signal === 'bearish' ? 'text-red-600' :
                                  'text-gray-600'
                                }`}>
                                  {signals.overall_signal?.toUpperCase() || 'N/A'}
                                </span>
                              </div>
                              <div className="flex justify-between">
                                <span className="text-gray-600 dark:text-gray-400">Confidence:</span>
                                <span className="text-blue-600 font-medium">
                                  {typeof signals.confidence === 'number' ? signals.confidence.toFixed(0) + '%' : 'N/A'}
                                </span>
                              </div>
                            </div>
                          )}

                          {/* Market Phase & Regime */}
                          {analysis && (
                            <div className="grid grid-cols-2 gap-2 text-xs">
                              {analysis.market_phase && (
                                <div className="flex justify-between">
                                  <span className="text-gray-600 dark:text-gray-400">Phase:</span>
                                  <span className="text-purple-600 font-medium">
                                    {analysis.market_phase.replace('_', ' ').toUpperCase()}
                                  </span>
                                </div>
                              )}
                              {analysis.market_regime && (
                                <div className="flex justify-between">
                                  <span className="text-gray-600 dark:text-gray-400">Regime:</span>
                                  <span className="text-indigo-600 font-medium">
                                    {analysis.market_regime.toUpperCase()}
                                  </span>
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      );
                    })()}
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

      {/* Dedicated Momentum Indicators Section */}
      <div className="bg-card rounded-lg border border-border p-6">
        <h3 className="text-lg font-semibold text-card-foreground mb-4">
          ⚡ Momentum Indicators Analysis
          <span className="text-sm text-muted-foreground ml-2">
            (Real-time Momentum Tracking)
          </span>
        </h3>
        
        {enrichedCandles.length > 0 ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
            {/* RSI Panel */}
            <div className="bg-muted/30 rounded-lg p-4 border border-border/50">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-card-foreground">RSI (Relative Strength Index)</h4>
                <div className="text-xs text-muted-foreground">14-period</div>
              </div>
              
              {enrichedCandles.slice(-1).map((candle, index) => {
                const indicators = (candle as any).data?.indicators;
                const rsi = indicators?.rsi;
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                return (
                  <div key={`rsi-${candle.timestamp}-${index}`} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">Current RSI:</span>
                      <span className={`text-lg font-bold ${
                        rsi && rsi > 70 ? 'text-red-600' : 
                        rsi && rsi < 30 ? 'text-green-600' : 
                        'text-yellow-600'
                      }`}>
                        {rsi ? rsi.toFixed(1) : 'N/A'}
                      </span>
                    </div>
                    
                    {rsi && (
                      <>
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                          <div 
                            className={`h-3 rounded-full transition-all duration-300 ${
                              rsi > 70 ? 'bg-red-500' : 
                              rsi < 30 ? 'bg-green-500' : 
                              'bg-yellow-500'
                            }`}
                            style={{ width: `${rsi}%` }}
                          />
                        </div>
                        
                        <div className="flex justify-between text-xs text-muted-foreground">
                          <span>Oversold (30)</span>
                          <span>Neutral (50)</span>
                          <span>Overbought (70)</span>
                        </div>
                        
                        <div className="text-center">
                          <span className={`text-sm font-medium px-2 py-1 rounded ${
                            rsi > 70 ? 'bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-400' : 
                            rsi < 30 ? 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400' : 
                            'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400'
                          }`}>
                            {rsi > 70 ? '🔴 OVERBOUGHT' : rsi < 30 ? '🟢 OVERSOLD' : '🟡 NEUTRAL'}
                          </span>
                        </div>
                      </>
                    )}
                    
                    <div className="text-xs text-muted-foreground text-center">
                      Updated: {timestamp}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* MACD Panel */}
            <div className="bg-muted/30 rounded-lg p-4 border border-border/50">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-card-foreground">MACD</h4>
                <div className="text-xs text-muted-foreground">12,26,9</div>
              </div>
              
              {enrichedCandles.slice(-1).map((candle, index) => {
                const indicators = (candle as any).data?.indicators;
                const macdData = indicators?.macd;
                const macdLine = macdData?.line;
                const macdSignal = macdData?.signal;
                const histogram = macdData?.histogram;
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                // Determine MACD trading signals
                let tradingSignal = 'Neutral';
                let signalColor = 'text-gray-600';
                
                if (macdLine !== undefined && macdSignal !== undefined && histogram !== undefined) {
                  const isAboveZero = macdLine > 0;
                  const isBullishCrossover = macdLine > macdSignal && histogram > 0;
                  const isBearishCrossover = macdLine < macdSignal && histogram < 0;
                  const isZeroCrossover = Math.abs(macdLine) < 0.001; // Near zero
                  
                  if (isBullishCrossover && isAboveZero) {
                    tradingSignal = 'Strong Bullish Signal';
                    signalColor = 'text-green-600';
                  } else if (isBullishCrossover && !isAboveZero) {
                    tradingSignal = 'Bullish Crossover';
                    signalColor = 'text-green-600';
                  } else if (isBearishCrossover && !isAboveZero) {
                    tradingSignal = 'Strong Bearish Signal';
                    signalColor = 'text-red-600';
                  } else if (isBearishCrossover && isAboveZero) {
                    tradingSignal = 'Bearish Crossover';
                    signalColor = 'text-red-600';
                  } else if (isZeroCrossover && histogram > 0) {
                    tradingSignal = 'Zero Line Crossover (Bull)';
                    signalColor = 'text-green-600';
                  } else if (isZeroCrossover && histogram < 0) {
                    tradingSignal = 'Zero Line Crossover (Bear)';
                    signalColor = 'text-red-600';
                  } else if (macdLine > macdSignal) {
                    tradingSignal = 'Bullish Momentum';
                    signalColor = 'text-green-600';
                  } else if (macdLine < macdSignal) {
                    tradingSignal = 'Bearish Momentum';
                    signalColor = 'text-red-600';
                  }
                }
                
                return (
                  <div key={`macd-${candle.timestamp}-${index}`} className="space-y-3">
                    <div className="grid grid-cols-2 gap-2 text-sm">
                      <div>
                        <span className="text-muted-foreground">MACD Line:</span>
                        <div className={`font-mono ${macdLine && macdLine > 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {macdLine ? macdLine.toFixed(4) : 'N/A'}
                        </div>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Signal Line:</span>
                        <div className="font-mono text-blue-600">
                          {macdSignal ? macdSignal.toFixed(4) : 'N/A'}
                        </div>
                      </div>
                    </div>
                    
                    {histogram !== undefined && (
                      <div>
                        <span className="text-muted-foreground text-sm">Histogram:</span>
                        <div className="flex items-center gap-2">
                          <div className={`font-mono text-sm ${histogram > 0 ? 'text-green-600' : 'text-red-600'}`}>
                            {histogram.toFixed(4)}
                          </div>
                          <div className={`w-4 h-4 rounded ${histogram > 0 ? 'bg-green-500' : 'bg-red-500'}`} />
                        </div>
                        
                        {/* Histogram Trend Bar */}
                        <div className="mt-1 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div 
                            className={`h-2 rounded-full transition-all duration-300 ${
                              histogram > 0 ? 'bg-green-500' : 'bg-red-500'
                            }`}
                            style={{ 
                              width: `${Math.min(Math.abs(histogram) * 1000, 100)}%`,
                              marginLeft: histogram < 0 ? `${Math.max(0, 100 - Math.abs(histogram) * 1000)}%` : '0'
                            }}
                          />
                        </div>
                      </div>
                    )}
                    
                    {/* Trading Signal Display */}
                    <div className="text-center space-y-2">
                      <div className="text-muted-foreground text-xs">Trading Signal:</div>
                      <span className={`text-sm font-medium px-3 py-2 rounded-lg border ${
                        tradingSignal.includes('Strong Bullish') ? 
                          'bg-green-100 text-green-800 border-green-300 dark:bg-green-900/30 dark:text-green-300 dark:border-green-700' : 
                        tradingSignal.includes('Bullish') || tradingSignal.includes('Bull') ? 
                          'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800' : 
                        tradingSignal.includes('Strong Bearish') ? 
                          'bg-red-100 text-red-800 border-red-300 dark:bg-red-900/30 dark:text-red-300 dark:border-red-700' :
                        tradingSignal.includes('Bearish') || tradingSignal.includes('Bear') ? 
                          'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800' :
                          'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-900/20 dark:text-gray-400 dark:border-gray-800'
                      }`}>
                        {tradingSignal.includes('Strong Bullish') ? '🚀' :
                         tradingSignal.includes('Bullish') || tradingSignal.includes('Bull') ? '🟢' : 
                         tradingSignal.includes('Strong Bearish') ? '📉' :
                         tradingSignal.includes('Bearish') || tradingSignal.includes('Bear') ? '🔴' : '🟡'} {tradingSignal}
                      </span>
                      
                      {/* Signal Strength Indicator */}
                      {macdLine !== undefined && macdSignal !== undefined && (
                        <div className="text-xs">
                          <span className="text-muted-foreground">Strength: </span>
                          <span className={`font-medium ${
                            Math.abs(macdLine - macdSignal) > 0.01 ? 'text-orange-600' : 'text-gray-600'
                          }`}>
                            {Math.abs(macdLine - macdSignal) > 0.01 ? 'Strong' : 'Weak'}
                          </span>
                        </div>
                      )}
                    </div>
                    
                    <div className="text-xs text-muted-foreground text-center">
                      Updated: {timestamp}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Williams %R Panel */}
            <div className="bg-muted/30 rounded-lg p-4 border border-border/50">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-card-foreground">Williams %R</h4>
                <div className="text-xs text-muted-foreground">14-period</div>
              </div>
              
              {enrichedCandles.slice(-1).map((candle, index) => {
                const indicators = (candle as any).data?.indicators;
                const williamsR = indicators?.williams_r;
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                return (
                  <div key={`williams-${candle.timestamp}-${index}`} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">Williams %R:</span>
                      <span className={`text-lg font-bold ${
                        williamsR && williamsR > -20 ? 'text-red-600' : 
                        williamsR && williamsR < -80 ? 'text-green-600' : 
                        'text-yellow-600'
                      }`}>
                        {williamsR ? williamsR.toFixed(1) : 'N/A'}
                      </span>
                    </div>
                    
                    {williamsR && (
                      <>
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                          <div 
                            className={`h-3 rounded-full transition-all duration-300 ${
                              williamsR > -20 ? 'bg-red-500' : 
                              williamsR < -80 ? 'bg-green-500' : 
                              'bg-yellow-500'
                            }`}
                            style={{ width: `${Math.abs(williamsR)}%` }}
                          />
                        </div>
                        
                        <div className="flex justify-between text-xs text-muted-foreground">
                          <span>Oversold (-80)</span>
                          <span>Neutral (-50)</span>
                          <span>Overbought (-20)</span>
                        </div>
                        
                        <div className="text-center">
                          <span className={`text-sm font-medium px-2 py-1 rounded ${
                            williamsR > -20 ? 'bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-400' : 
                            williamsR < -80 ? 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400' : 
                            'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400'
                          }`}>
                            {williamsR > -20 ? '🔴 OVERBOUGHT' : williamsR < -80 ? '🟢 OVERSOLD' : '🟡 NEUTRAL'}
                          </span>
                        </div>
                      </>
                    )}
                    
                    <div className="text-xs text-muted-foreground text-center">
                      Updated: {timestamp}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ) : (
          <div className="h-48 flex items-center justify-center">
            <div className="text-center">
              <div className="text-4xl mb-4">⚡</div>
              <p className="text-lg text-muted-foreground mb-2">
                Waiting for momentum data...
              </p>
              <p className="text-sm text-muted-foreground">
                Momentum indicators will display here in real-time
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Dedicated Volatility Indicators Section */}
      <div className="bg-card rounded-lg border border-border p-6">
        <h3 className="text-lg font-semibold text-card-foreground mb-4">
          📊 Volatility Indicators Analysis
          <span className="text-sm text-muted-foreground ml-2">
            (Real-time Volatility Tracking)
          </span>
        </h3>
        
        {enrichedCandles.length > 0 ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
            {/* ATR Panel */}
            <div className="bg-muted/30 rounded-lg p-4 border border-border/50">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-card-foreground">ATR (Average True Range)</h4>
                <div className="text-xs text-muted-foreground">14-period</div>
              </div>
              
              {enrichedCandles.slice(-1).map((candle, index) => {
                const indicators = (candle as any).data?.indicators;
                const analysis = (candle as any).data?.analysis;
                const atr = indicators?.atr;
                const volatilityLevel = analysis?.volatility_level;
                const volatilityPercent = analysis?.volatility_percent;
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                return (
                  <div key={`atr-${candle.timestamp}-${index}`} className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">ATR Value:</span>
                      <span className="text-lg font-bold text-orange-600">
                        {atr ? atr.toFixed(3) : 'N/A'}
                      </span>
                    </div>
                    
                    {volatilityPercent && (
                      <div className="space-y-2">
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                          <div 
                            className={`h-3 rounded-full transition-all duration-300 ${
                              volatilityLevel === 'high' ? 'bg-red-500' : 
                              volatilityLevel === 'low' ? 'bg-green-500' : 
                              'bg-yellow-500'
                            }`}
                            style={{ width: `${Math.min(volatilityPercent, 100)}%` }}
                          />
                        </div>
                        
                        <div className="flex justify-between text-xs text-muted-foreground">
                          <span>Low (0%)</span>
                          <span>Normal (50%)</span>
                          <span>High (100%)</span>
                        </div>
                      </div>
                    )}
                    
                    {volatilityLevel && (
                      <div className="text-center">
                        <span className={`text-sm font-medium px-2 py-1 rounded ${
                          volatilityLevel === 'high' ? 'bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-400' : 
                          volatilityLevel === 'low' ? 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400' : 
                          'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400'
                        }`}>
                          {volatilityLevel === 'high' ? '🔴 HIGH VOLATILITY' : 
                           volatilityLevel === 'low' ? '🟢 LOW VOLATILITY' : 
                           '🟡 NORMAL VOLATILITY'}
                        </span>
                      </div>
                    )}
                    
                    {volatilityPercent && (
                      <div className="text-center text-sm">
                        <span className="text-muted-foreground">Volatility Level: </span>
                        <span className="font-medium text-orange-600">{volatilityPercent.toFixed(0)}%</span>
                      </div>
                    )}
                    
                    <div className="text-xs text-muted-foreground text-center">
                      Updated: {timestamp}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Bollinger Bands Panel */}
            <div className="bg-muted/30 rounded-lg p-4 border border-border/50">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-card-foreground">Bollinger Bands</h4>
                <div className="text-xs text-muted-foreground">20,2</div>
              </div>
              
              {enrichedCandles.slice(-1).map((candle, index) => {
                const indicators = (candle as any).data?.indicators;
                const bollingerBands = indicators?.bollinger_bands;
                const upperBand = bollingerBands?.upper;
                const lowerBand = bollingerBands?.lower;
                const middleBand = bollingerBands?.middle || indicators?.sma_20;
                const currentPrice = candle.close;
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                // Calculate position within bands
                let bandPosition = 'middle';
                let bandPercent = 50;
                if (upperBand && lowerBand && currentPrice) {
                  const bandWidth = upperBand - lowerBand;
                  const priceFromLower = currentPrice - lowerBand;
                  bandPercent = (priceFromLower / bandWidth) * 100;
                  
                  if (bandPercent > 80) bandPosition = 'upper';
                  else if (bandPercent < 20) bandPosition = 'lower';
                  else bandPosition = 'middle';
                }
                
                return (
                  <div key={`bb-${candle.timestamp}-${index}`} className="space-y-3">
                    <div className="grid grid-cols-2 gap-2 text-sm">
                      <div>
                        <span className="text-muted-foreground">Upper:</span>
                        <div className="font-mono text-red-600">
                          ${upperBand ? upperBand.toFixed(2) : 'N/A'}
                        </div>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Lower:</span>
                        <div className="font-mono text-green-600">
                          ${lowerBand ? lowerBand.toFixed(2) : 'N/A'}
                        </div>
                      </div>
                    </div>
                    
                    {middleBand && (
                      <div className="text-center">
                        <span className="text-muted-foreground text-sm">Middle (SMA20):</span>
                        <div className="font-mono text-blue-600">
                          ${middleBand.toFixed(2)}
                        </div>
                      </div>
                    )}
                    
                    {upperBand && lowerBand && (
                      <>
                        <div className="space-y-2">
                          <div className="text-center text-sm">
                            <span className="text-muted-foreground">Price Position: </span>
                            <span className={`font-medium ${
                              bandPosition === 'upper' ? 'text-red-600' :
                              bandPosition === 'lower' ? 'text-green-600' :
                              'text-blue-600'
                            }`}>
                              {bandPercent.toFixed(0)}%
                            </span>
                          </div>
                          
                          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3 relative">
                            <div 
                              className="h-3 bg-blue-500 rounded-full transition-all duration-300"
                              style={{ width: `${Math.max(0, Math.min(100, bandPercent))}%` }}
                            />
                            <div className="absolute inset-0 flex justify-between items-center px-1">
                              <div className="w-1 h-1 bg-green-600 rounded-full" title="Lower Band" />
                              <div className="w-1 h-1 bg-red-600 rounded-full" title="Upper Band" />
                            </div>
                          </div>
                        </div>
                        
                        <div className="text-center">
                          <span className={`text-sm font-medium px-2 py-1 rounded ${
                            bandPosition === 'upper' ? 'bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-400' : 
                            bandPosition === 'lower' ? 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400' : 
                            'bg-blue-100 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400'
                          }`}>
                            {bandPosition === 'upper' ? '🔴 NEAR UPPER BAND' : 
                             bandPosition === 'lower' ? '🟢 NEAR LOWER BAND' : 
                             '🔵 MIDDLE RANGE'}
                          </span>
                        </div>
                        
                        <div className="text-xs text-center">
                          <span className="text-muted-foreground">Band Width: </span>
                          <span className="text-pink-600 font-medium">
                            ${(upperBand - lowerBand).toFixed(2)}
                          </span>
                        </div>
                      </>
                    )}
                    
                    <div className="text-xs text-muted-foreground text-center">
                      Updated: {timestamp}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Volatility Summary Panel */}
            <div className="bg-muted/30 rounded-lg p-4 border border-border/50">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-card-foreground">Volatility Summary</h4>
                <div className="text-xs text-muted-foreground">Combined</div>
              </div>
              
              {enrichedCandles.slice(-1).map((candle, index) => {
                const indicators = (candle as any).data?.indicators;
                const analysis = (candle as any).data?.analysis;
                const atr = indicators?.atr;
                const volatilityLevel = analysis?.volatility_level;
                const volatilityPercent = analysis?.volatility_percent;
                const bollingerBands = indicators?.bollinger_bands;
                const upperBand = bollingerBands?.upper;
                const lowerBand = bollingerBands?.lower;
                const timestamp = new Date(candle.timestamp).toLocaleTimeString();
                
                // Calculate price range from yesterday
                const priceRange = candle.high - candle.low;
                const priceRangePercent = (priceRange / candle.open) * 100;
                
                return (
                  <div key={`vol-summary-${candle.timestamp}-${index}`} className="space-y-3">
                    {/* Overall Volatility Status */}
                    <div className="text-center p-3 rounded-lg bg-muted/50 border border-border/30">
                      <div className="text-lg font-bold mb-2">
                        {volatilityLevel === 'high' ? '🔥' : volatilityLevel === 'low' ? '😴' : '📊'}
                      </div>
                      <div className={`text-sm font-semibold ${
                        volatilityLevel === 'high' ? 'text-red-600' : 
                        volatilityLevel === 'low' ? 'text-green-600' : 
                        'text-yellow-600'
                      }`}>
                        {volatilityLevel ? volatilityLevel.toUpperCase() : 'NORMAL'} VOLATILITY
                      </div>
                      {volatilityPercent && (
                        <div className="text-xs text-muted-foreground mt-1">
                          {volatilityPercent.toFixed(0)}% intensity
                        </div>
                      )}
                    </div>
                    
                    {/* Key Metrics */}
                    <div className="space-y-2 text-sm">
                      {atr && (
                        <div className="flex justify-between items-center">
                          <span className="text-muted-foreground">ATR:</span>
                          <span className="font-mono text-orange-600">{atr.toFixed(3)}</span>
                        </div>
                      )}
                      
                      <div className="flex justify-between items-center">
                        <span className="text-muted-foreground">Today's Range:</span>
                        <span className="font-mono text-purple-600">
                          ${priceRange.toFixed(2)} ({priceRangePercent.toFixed(1)}%)
                        </span>
                      </div>
                      
                      {upperBand && lowerBand && (
                        <div className="flex justify-between items-center">
                          <span className="text-muted-foreground">BB Width:</span>
                          <span className="font-mono text-pink-600">
                            ${(upperBand - lowerBand).toFixed(2)}
                          </span>
                        </div>
                      )}
                    </div>
                    
                    {/* Volatility Interpretation */}
                    <div className="p-2 rounded bg-muted/30 border border-border/20">
                      <div className="text-xs text-muted-foreground mb-1">Market State:</div>
                      <div className="text-sm">
                        {volatilityLevel === 'high' ? (
                          <span className="text-red-600">⚠️ High volatility - Expect larger price swings</span>
                        ) : volatilityLevel === 'low' ? (
                          <span className="text-green-600">✅ Low volatility - More stable price action</span>
                        ) : (
                          <span className="text-yellow-600">📊 Normal volatility - Typical price movement</span>
                        )}
                      </div>
                    </div>
                    
                    <div className="text-xs text-muted-foreground text-center">
                      Updated: {timestamp}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ) : (
          <div className="h-48 flex items-center justify-center">
            <div className="text-center">
              <div className="text-4xl mb-4">📊</div>
              <p className="text-lg text-muted-foreground mb-2">
                Waiting for volatility data...
              </p>
              <p className="text-sm text-muted-foreground">
                Volatility indicators will display here in real-time
              </p>
            </div>
          </div>
        )}
      </div>

    </div>
  );
};

export default Dashboard;
