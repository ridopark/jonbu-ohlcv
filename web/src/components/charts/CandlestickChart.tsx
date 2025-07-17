import React from 'react';
import { useChartStore } from '../../stores/chartStore';

const CandlestickChart: React.FC = () => {
  const { 
    candles, 
    enrichedCandles, 
    symbol, 
    timeframe, 
    supportResistanceLevels, 
    showSupportResistance 
  } = useChartStore();
  
  // Debug logging
  React.useEffect(() => {
    console.log('📊 Chart component state update:', {
      symbol,
      timeframe,
      candlesCount: candles.length,
      enrichedCandlesCount: enrichedCandles.length,
      firstCandle: candles[0],
      lastCandle: candles[candles.length - 1]
    });
  }, [candles, enrichedCandles, symbol, timeframe]);
  
  // For now, we'll create a simple placeholder chart
  // In a real implementation, this would use a charting library like Recharts or Chart.js
  
  const chartData = React.useMemo(() => {
    if (candles.length === 0) return [];
    
    // Filter out invalid candles and ensure numeric values
    const validCandles = candles.filter(candle => 
      candle &&
      typeof candle.open === 'number' && !isNaN(candle.open) &&
      typeof candle.high === 'number' && !isNaN(candle.high) &&
      typeof candle.low === 'number' && !isNaN(candle.low) &&
      typeof candle.close === 'number' && !isNaN(candle.close) &&
      typeof candle.volume === 'number' && !isNaN(candle.volume) &&
      candle.open > 0 && candle.high > 0 && candle.low > 0 && candle.close > 0
    );
    
    return validCandles.slice(-100).map((candle, index) => ({
      ...candle,
      index,
      enriched: enrichedCandles.find(e => e.timestamp === candle.timestamp),
    }));
  }, [candles, enrichedCandles]);
  
  const priceRange = React.useMemo(() => {
    if (chartData.length === 0) return { min: 0, max: 100 };
    
    const allPrices = chartData.flatMap(d => [d.high, d.low]).filter(price => 
      typeof price === 'number' && !isNaN(price) && price > 0
    );
    
    if (allPrices.length === 0) return { min: 0, max: 100 };
    
    const min = Math.min(...allPrices);
    const max = Math.max(...allPrices);
    const range = max - min;
    
    // Prevent division by zero and ensure valid range
    if (range === 0 || !isFinite(range)) {
      return { min: min - 1, max: min + 1 };
    }
    
    const padding = range * 0.1;
    
    return {
      min: Math.max(0, min - padding),
      max: max + padding,
    };
  }, [chartData]);
  
  const volumeRange = React.useMemo(() => {
    if (chartData.length === 0) return { min: 0, max: 1000 };
    
    const volumes = chartData.map(d => d.volume).filter(vol => 
      typeof vol === 'number' && !isNaN(vol) && vol >= 0
    );
    
    if (volumes.length === 0) return { min: 0, max: 1000 };
    
    const max = Math.max(...volumes);
    
    return {
      min: 0,
      max: Math.max(max * 1.2, 1), // Ensure minimum max value
    };
  }, [chartData]);
   const formatPrice = (price: number) => {
    if (!isFinite(price) || isNaN(price)) return '$0.00';
    return `$${price.toFixed(2)}`;
  };

  const formatVolume = (volume: number) => {
    if (!isFinite(volume) || isNaN(volume) || volume < 0) return '0';
    if (volume >= 1e6) return `${(volume / 1e6).toFixed(1)}M`;
    if (volume >= 1e3) return `${(volume / 1e3).toFixed(1)}K`;
    return volume.toString();
  };

  if (chartData.length === 0) {
    return (
      <div className="h-96 flex items-center justify-center bg-muted rounded-lg">
        <div className="text-center">
          <div className="text-4xl mb-2">📈</div>
          <h3 className="text-lg font-medium text-muted-foreground">No Data Available</h3>
          <p className="text-sm text-muted-foreground">
            Start streaming data for {symbol} to see the chart
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Chart Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-card-foreground">
            {symbol} Candlestick Chart
          </h3>
          <p className="text-sm text-muted-foreground">
            Timeframe: {timeframe} • {chartData.length} candles
          </p>
        </div>
        
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 bg-green-500 rounded" />
            <span>Bullish</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 bg-red-500 rounded" />
            <span>Bearish</span>
          </div>
          {enrichedCandles.length > 0 && (
            <>
              <div className="flex items-center gap-1 ml-4">
                <div className="w-4 h-0.5 bg-blue-500" />
                <span>SMA 20</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-4 h-0.5 bg-purple-500" />
                <span>SMA 50</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-4 h-0.5 bg-orange-500" />
                <span>RSI</span>
              </div>
            </>
          )}
          {showSupportResistance && supportResistanceLevels && (
            <>
              <div className="flex items-center gap-1 ml-4">
                <div className="w-4 h-0.5 bg-green-500" />
                <span>Support</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-4 h-0.5 bg-red-500" />
                <span>Resistance</span>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Main Chart Area */}
      <div className="relative bg-background border border-border rounded-lg p-4">
        <div className="h-96 relative overflow-hidden">
          {/* Price Chart */}
          <svg
            width="100%"
            height="240"
            viewBox="0 0 1000 240"
            className="absolute top-0 left-0"
            preserveAspectRatio="none"
          >
            {/* Grid lines */}
            <defs>
              <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
                <path d="M 20 0 L 0 0 0 20" fill="none" stroke="rgb(var(--color-border))" strokeWidth="0.5" opacity="0.3"/>
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#grid)" />
            
            {/* Candlesticks */}
            {chartData.map((candle, index) => {
              const x = (index / (chartData.length - 1)) * 980 + 10;
              const isGreen = candle.close > candle.open;
              
              // Scale prices to chart height with NaN protection
              const scalePrice = (price: number) => {
                if (!isFinite(price) || isNaN(price)) return 110; // Middle of chart
                const range = priceRange.max - priceRange.min;
                if (range === 0) return 110;
                return 220 - ((price - priceRange.min) / range) * 200;
              };
              
              const highY = scalePrice(candle.high);
              const lowY = scalePrice(candle.low);
              const openY = scalePrice(candle.open);
              const closeY = scalePrice(candle.close);
              
              // Additional NaN checks
              if (!isFinite(x) || !isFinite(highY) || !isFinite(lowY) || !isFinite(openY) || !isFinite(closeY)) {
                return null;
              }
              
              const bodyTop = Math.min(openY, closeY);
              const bodyHeight = Math.abs(openY - closeY);
              
              return (
                <g key={`candle-${index}`}>
                  {/* Wick */}
                  <line
                    x1={x}
                    y1={highY}
                    x2={x}
                    y2={lowY}
                    stroke={isGreen ? 'rgb(16, 185, 129)' : 'rgb(239, 68, 68)'}
                    strokeWidth="1"
                  />
                  
                  {/* Body */}
                  <rect
                    x={x - 3}
                    y={bodyTop}
                    width="6"
                    height={Math.max(bodyHeight, 1)}
                    fill={isGreen ? 'rgb(16, 185, 129)' : 'rgb(239, 68, 68)'}
                    stroke={isGreen ? 'rgb(16, 185, 129)' : 'rgb(239, 68, 68)'}
                    strokeWidth="1"
                  />
                </g>
              );
            }).filter(Boolean)}
            
            {/* Technical Indicator Lines */}
            {enrichedCandles.length > 1 && (() => {
              // Scale price function for indicators
              const scalePrice = (price: number) => {
                if (!isFinite(price) || isNaN(price)) return 110;
                const range = priceRange.max - priceRange.min;
                if (range === 0) return 110;
                return 220 - ((price - priceRange.min) / range) * 200;
              };
              
              // Get enriched candles that match our chart data timeframe
              const enrichedChartData = chartData.map(candle => 
                enrichedCandles.find(e => e.timestamp === candle.timestamp)
              ).filter(Boolean);
              
              if (enrichedChartData.length < 2) return null;
              
              // SMA 20 Line
              const sma20Points = enrichedChartData
                .map((enriched, index) => {
                  if (!enriched || typeof enriched.sma20 !== 'number') return null;
                  const x = (index / (chartData.length - 1)) * 980 + 10;
                  const y = scalePrice(enriched.sma20);
                  return { x, y };
                })
                .filter((p): p is { x: number; y: number } => p !== null);
              
              // SMA 50 Line  
              const sma50Points = enrichedChartData
                .map((enriched, index) => {
                  if (!enriched || typeof enriched.sma50 !== 'number') return null;
                  const x = (index / (chartData.length - 1)) * 980 + 10;
                  const y = scalePrice(enriched.sma50);
                  return { x, y };
                })
                .filter((p): p is { x: number; y: number } => p !== null);
              
              return (
                <g>
                  {/* SMA 20 Line */}
                  {sma20Points.length > 1 && (
                    <polyline
                      points={sma20Points.map(p => `${p.x},${p.y}`).join(' ')}
                      fill="none"
                      stroke="rgb(59, 130, 246)"
                      strokeWidth="2"
                      opacity="0.8"
                    />
                  )}
                  
                  {/* SMA 50 Line */}
                  {sma50Points.length > 1 && (
                    <polyline
                      points={sma50Points.map(p => `${p.x},${p.y}`).join(' ')}
                      fill="none"
                      stroke="rgb(147, 51, 234)"
                      strokeWidth="2"
                      opacity="0.8"
                    />
                  )}
                </g>
              );
            })()}
            
            {/* Support and Resistance Levels */}
            {showSupportResistance && supportResistanceLevels && (
              <g>
                {/* Support Lines */}
                {supportResistanceLevels.support?.map((level, index) => {
                  const y = 220 - ((level.price - priceRange.min) / (priceRange.max - priceRange.min)) * 200;
                  
                  // Skip invalid levels
                  if (!isFinite(y) || level.price < priceRange.min || level.price > priceRange.max) return null;
                  
                  // Calculate line opacity and thickness based on strength and confidence
                  const opacity = Math.min(0.9, Math.max(0.3, level.confidence / 100));
                  const strokeWidth = Math.max(1.5, Math.min(4, level.strength / 25));
                  
                  return (
                    <g key={`support-${index}`}>
                      {/* Support zone background */}
                      <rect
                        x="10"
                        y={y - 2}
                        width="980"
                        height="4"
                        fill="rgb(34, 197, 94)"
                        opacity={0.1}
                      />
                      
                      {/* Main support line */}
                      <line
                        x1="10"
                        y1={y}
                        x2="990"
                        y2={y}
                        stroke="rgb(34, 197, 94)"
                        strokeWidth={strokeWidth}
                        strokeDasharray={level.confidence > 70 ? "0" : "8,4"}
                        opacity={opacity}
                      />
                      
                      {/* Touch point indicators */}
                      {Array.from({ length: Math.min(level.touches, 5) }, (_, i) => (
                        <circle
                          key={`touch-${i}`}
                          cx={100 + i * 180}
                          cy={y}
                          r="3"
                          fill="rgb(34, 197, 94)"
                          opacity={0.8}
                        />
                      ))}
                      
                      {/* Support label with enhanced styling */}
                      <g>
                        <rect
                          x="880"
                          y={y - 20}
                          width="115"
                          height="16"
                          fill="rgb(34, 197, 94)"
                          opacity={0.9}
                          rx="2"
                        />
                        <text
                          x="888"
                          y={y - 8}
                          className="text-xs font-medium"
                          fill="white"
                        >
                          SUP {formatPrice(level.price)}
                        </text>
                        <text
                          x="888"
                          y={y + 14}
                          className="text-xs"
                          fill="rgb(34, 197, 94)"
                          opacity={0.8}
                        >
                          {level.touches}x • {level.confidence.toFixed(0)}%
                        </text>
                      </g>
                    </g>
                  );
                })}
                
                {/* Resistance Lines */}
                {supportResistanceLevels.resistance?.map((level, index) => {
                  const y = 220 - ((level.price - priceRange.min) / (priceRange.max - priceRange.min)) * 200;
                  
                  // Skip invalid levels
                  if (!isFinite(y) || level.price < priceRange.min || level.price > priceRange.max) return null;
                  
                  // Calculate line opacity and thickness based on strength and confidence
                  const opacity = Math.min(0.9, Math.max(0.3, level.confidence / 100));
                  const strokeWidth = Math.max(1.5, Math.min(4, level.strength / 25));
                  
                  return (
                    <g key={`resistance-${index}`}>
                      {/* Resistance zone background */}
                      <rect
                        x="10"
                        y={y - 2}
                        width="980"
                        height="4"
                        fill="rgb(239, 68, 68)"
                        opacity={0.1}
                      />
                      
                      {/* Main resistance line */}
                      <line
                        x1="10"
                        y1={y}
                        x2="990"
                        y2={y}
                        stroke="rgb(239, 68, 68)"
                        strokeWidth={strokeWidth}
                        strokeDasharray={level.confidence > 70 ? "0" : "8,4"}
                        opacity={opacity}
                      />
                      
                      {/* Touch point indicators */}
                      {Array.from({ length: Math.min(level.touches, 5) }, (_, i) => (
                        <circle
                          key={`touch-${i}`}
                          cx={100 + i * 180}
                          cy={y}
                          r="3"
                          fill="rgb(239, 68, 68)"
                          opacity={0.8}
                        />
                      ))}
                      
                      {/* Resistance label with enhanced styling */}
                      <g>
                        <rect
                          x="880"
                          y={y + 6}
                          width="115"
                          height="16"
                          fill="rgb(239, 68, 68)"
                          opacity={0.9}
                          rx="2"
                        />
                        <text
                          x="888"
                          y={y + 18}
                          className="text-xs font-medium"
                          fill="white"
                        >
                          RES {formatPrice(level.price)}
                        </text>
                        <text
                          x="888"
                          y={y - 2}
                          className="text-xs"
                          fill="rgb(239, 68, 68)"
                          opacity={0.8}
                        >
                          {level.touches}x • {level.confidence.toFixed(0)}%
                        </text>
                      </g>
                    </g>
                  );
                })}
                
                {/* Current price position indicator */}
                {supportResistanceLevels.current && (
                  <g>
                    {/* Position status banner */}
                    <rect
                      x="350"
                      y="5"
                      width="300"
                      height="20"
                      fill={
                        supportResistanceLevels.current.position === 'near_support' ? 'rgb(34, 197, 94)' :
                        supportResistanceLevels.current.position === 'near_resistance' ? 'rgb(239, 68, 68)' :
                        'rgb(59, 130, 246)'
                      }
                      opacity={0.1}
                      rx="4"
                    />
                    <text
                      x="500"
                      y="18"
                      className="text-sm font-medium"
                      fill={
                        supportResistanceLevels.current.position === 'near_support' ? 'rgb(34, 197, 94)' :
                        supportResistanceLevels.current.position === 'near_resistance' ? 'rgb(239, 68, 68)' :
                        'rgb(59, 130, 246)'
                      }
                      textAnchor="middle"
                    >
                      {supportResistanceLevels.current.position.replace('_', ' ').toUpperCase()}
                      {supportResistanceLevels.current.distance_to_support > 0 && 
                        ` (${supportResistanceLevels.current.distance_to_support.toFixed(1)}% from SUP)`}
                      {supportResistanceLevels.current.distance_to_resistance > 0 && 
                        ` (${supportResistanceLevels.current.distance_to_resistance.toFixed(1)}% to RES)`}
                    </text>
                  </g>
                )}
              </g>
            )}
            
            {/* Price labels */}
            <g className="text-xs fill-current text-muted-foreground">
              {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
                const price = priceRange.min + (priceRange.max - priceRange.min) * ratio;
                const y = 220 - ratio * 200;
                
                // Skip invalid price labels
                if (!isFinite(price) || !isFinite(y)) return null;
                
                return (
                  <text key={ratio} x="5" y={y + 4} className="text-xs">
                    {formatPrice(price)}
                  </text>
                );
              }).filter(Boolean)}
            </g>
          </svg>
          
          {/* Volume Chart */}
          <svg
            width="100%"
            height="60"
            viewBox="0 0 1000 60"
            className="absolute bottom-0 left-0"
            preserveAspectRatio="none"
          >
            {chartData.map((candle, index) => {
              const x = (index / (chartData.length - 1)) * 980 + 10;
              const height = (candle.volume / volumeRange.max) * 50;
              const isGreen = candle.close > candle.open;
              
              // NaN protection for volume chart
              if (!isFinite(x) || !isFinite(height) || height < 0) {
                return null;
              }
              
              return (
                <rect
                  key={`volume-${index}`}
                  x={x - 2}
                  y={50 - height}
                  width="4"
                  height={height}
                  fill={isGreen ? 'rgb(16, 185, 129)' : 'rgb(239, 68, 68)'}
                  opacity="0.6"
                />
              );
            }).filter(Boolean)}
            
            {/* Volume labels */}
            <text x="5" y="15" className="text-xs fill-current text-muted-foreground">
              {formatVolume(volumeRange.max)}
            </text>
            <text x="5" y="50" className="text-xs fill-current text-muted-foreground">
              0
            </text>
          </svg>
          
          {/* RSI Chart - Separate panel below main chart */}
          <svg
            width="100%"
            height="80"
            viewBox="0 0 1000 80"
            className="absolute bottom-16 left-0"
            preserveAspectRatio="none"
          >
            {/* RSI Background */}
            <rect width="100%" height="100%" fill="rgba(0,0,0,0.02)" />
            
            {/* RSI Overbought/Oversold lines */}
            <line x1="0" y1="16" x2="1000" y2="16" stroke="rgb(239, 68, 68)" strokeWidth="1" strokeDasharray="2,2" opacity="0.5" />
            <line x1="0" y1="64" x2="1000" y2="64" stroke="rgb(16, 185, 129)" strokeWidth="1" strokeDasharray="2,2" opacity="0.5" />
            <line x1="0" y1="40" x2="1000" y2="40" stroke="rgb(156, 163, 175)" strokeWidth="1" strokeDasharray="1,1" opacity="0.3" />
            
            {/* RSI Line */}
            {enrichedCandles.length > 1 && (() => {
              const enrichedChartData = chartData.map(candle => 
                enrichedCandles.find(e => e.timestamp === candle.timestamp)
              ).filter(Boolean);
              
              if (enrichedChartData.length < 2) return null;
              
              const rsiPoints = enrichedChartData
                .map((enriched, index) => {
                  if (!enriched || typeof enriched.rsi !== 'number') return null;
                  const x = (index / (chartData.length - 1)) * 980 + 10;
                  const y = 80 - (enriched.rsi / 100) * 80; // Scale RSI 0-100 to chart height
                  return { x, y };
                })
                .filter((p): p is { x: number; y: number } => p !== null);
              
              return rsiPoints.length > 1 ? (
                <polyline
                  points={rsiPoints.map(p => `${p.x},${p.y}`).join(' ')}
                  fill="none"
                  stroke="rgb(251, 146, 60)"
                  strokeWidth="2"
                  opacity="0.9"
                />
              ) : null;
            })()}
            
            {/* RSI Labels */}
            <text x="5" y="12" className="text-xs fill-current text-red-600">70</text>
            <text x="5" y="36" className="text-xs fill-current text-muted-foreground">50</text>
            <text x="5" y="68" className="text-xs fill-current text-green-600">30</text>
            <text x="5" y="78" className="text-xs fill-current text-muted-foreground">RSI</text>
          </svg>
        </div>
        
        {/* Chart Footer */}
        <div className="mt-2 flex justify-between text-xs text-muted-foreground">
          <span>Volume</span>
          <span>Price: {formatPrice(priceRange.min)} - {formatPrice(priceRange.max)}</span>
        </div>
      </div>
    </div>
  );
};

export default CandlestickChart;
