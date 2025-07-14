import React from 'react';
import { useChartStore } from '../../stores/chartStore';

const StatsCards: React.FC = () => {
  const { candles, enrichedCandles, symbol, isStreaming } = useChartStore();
  
  // Calculate stats from latest candles
  const stats = React.useMemo(() => {
    if (candles.length === 0) {
      return {
        currentPrice: 0,
        priceChange: 0,
        priceChangePercent: 0,
        volume: 0,
        high24h: 0,
        low24h: 0,
      };
    }
    
    const latest = candles[candles.length - 1];
    const previous = candles.length > 1 ? candles[candles.length - 2] : latest;
    
    const priceChange = latest.close - previous.close;
    const priceChangePercent = previous.close !== 0 ? (priceChange / previous.close) * 100 : 0;
    
    // Calculate 24h high/low from recent candles
    const recentCandles = candles.slice(-1440); // Approximate 24h for 1m candles
    const high24h = Math.max(...recentCandles.map(c => c.high));
    const low24h = Math.min(...recentCandles.map(c => c.low));
    
    return {
      currentPrice: latest.close,
      priceChange,
      priceChangePercent,
      volume: latest.volume,
      high24h,
      low24h,
    };
  }, [candles]);
  
  const formatPrice = (price: number) => {
    return price.toLocaleString('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  };
  
  const formatVolume = (volume: number) => {
    if (volume >= 1e9) {
      return `${(volume / 1e9).toFixed(1)}B`;
    }
    if (volume >= 1e6) {
      return `${(volume / 1e6).toFixed(1)}M`;
    }
    if (volume >= 1e3) {
      return `${(volume / 1e3).toFixed(1)}K`;
    }
    return volume.toLocaleString();
  };

  const cards = [
    {
      title: 'Current Price',
      value: formatPrice(stats.currentPrice),
      subtitle: `${symbol}`,
      icon: '💰',
      color: 'text-card-foreground',
    },
    {
      title: '24h Change',
      value: formatPrice(Math.abs(stats.priceChange)),
      subtitle: `${stats.priceChangePercent >= 0 ? '+' : ''}${stats.priceChangePercent.toFixed(2)}%`,
      icon: stats.priceChangePercent >= 0 ? '📈' : '📉',
      color: stats.priceChangePercent >= 0 ? 'text-green-600' : 'text-red-600',
    },
    {
      title: 'Volume',
      value: formatVolume(stats.volume),
      subtitle: 'Current candle',
      icon: '📊',
      color: 'text-card-foreground',
    },
    {
      title: '24h High',
      value: formatPrice(stats.high24h),
      subtitle: 'Daily high',
      icon: '⬆️',
      color: 'text-green-600',
    },
    {
      title: '24h Low',
      value: formatPrice(stats.low24h),
      subtitle: 'Daily low',
      icon: '⬇️',
      color: 'text-red-600',
    },
    {
      title: 'Data Points',
      value: candles.length.toLocaleString(),
      subtitle: `${enrichedCandles.length} enriched`,
      icon: '🔢',
      color: 'text-card-foreground',
    },
  ];

  return (
    <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
      {cards.map((card, index) => (
        <div
          key={card.title}
          className="bg-card rounded-lg border border-border p-4 hover:shadow-md transition-shadow"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-2xl">{card.icon}</span>
            {isStreaming && index === 0 && (
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            )}
          </div>
          
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
              {card.title}
            </p>
            <p className={`text-lg font-bold ${card.color}`}>
              {card.value}
            </p>
            <p className="text-xs text-muted-foreground">
              {card.subtitle}
            </p>
          </div>
        </div>
      ))}
    </div>
  );
};

export default StatsCards;
