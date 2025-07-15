import React from 'react';
import { useChartStore } from '../../stores/chartStore';

const CandlestickChart: React.FC = () => {
  const { candles, enrichedCandles, symbol, timeframe, showIndicators } = useChartStore();
  
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
        </div>
      </div>

      {/* Main Chart Area */}
      <div className="relative bg-background border border-border rounded-lg p-4">
        <div className="h-80 relative overflow-hidden">
          {/* Price Chart */}
          <svg
            width="100%"
            height="240"
            viewBox="0 0 800 240"
            className="absolute top-0 left-0"
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
              const x = (index / (chartData.length - 1)) * 780 + 10;
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
            viewBox="0 0 800 60"
            className="absolute bottom-0 left-0"
          >
            {chartData.map((candle, index) => {
              const x = (index / (chartData.length - 1)) * 780 + 10;
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
        </div>
        
        {/* Chart Footer */}
        <div className="mt-2 flex justify-between text-xs text-muted-foreground">
          <span>Volume</span>
          <span>Price: {formatPrice(priceRange.min)} - {formatPrice(priceRange.max)}</span>
        </div>
      </div>
      
      {/* Technical Indicators */}
      {showIndicators && enrichedCandles.length > 0 && (
        <div className="bg-card border border-border rounded-lg p-4">
          <h4 className="text-sm font-medium text-card-foreground mb-3">
            Technical Indicators
          </h4>
          
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
            {(() => {
              const latest = enrichedCandles[enrichedCandles.length - 1];
              if (!latest) return null;
              
              return (
                <>
                  {latest.sma20 && (
                    <div>
                      <span className="text-muted-foreground">SMA 20:</span>
                      <span className="ml-2 font-medium">{formatPrice(latest.sma20)}</span>
                    </div>
                  )}
                  
                  {latest.ema50 && (
                    <div>
                      <span className="text-muted-foreground">EMA 50:</span>
                      <span className="ml-2 font-medium">{formatPrice(latest.ema50)}</span>
                    </div>
                  )}
                  
                  {latest.rsi && (
                    <div>
                      <span className="text-muted-foreground">RSI:</span>
                      <span className={`ml-2 font-medium ${
                        latest.rsi > 70 ? 'text-red-600' : 
                        latest.rsi < 30 ? 'text-green-600' : 
                        'text-card-foreground'
                      }`}>
                        {latest.rsi.toFixed(1)}
                      </span>
                    </div>
                  )}
                  
                  {latest.macd && (
                    <div>
                      <span className="text-muted-foreground">MACD:</span>
                      <span className="ml-2 font-medium">{latest.macd.toFixed(4)}</span>
                    </div>
                  )}
                </>
              );
            })()}
          </div>
        </div>
      )}
    </div>
  );
};

export default CandlestickChart;
