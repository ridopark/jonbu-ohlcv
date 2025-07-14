import React from 'react';
import { useChartStore, type Timeframe } from '../../stores/chartStore';

const TimeframeSelector: React.FC = () => {
  const { timeframe, setTimeframe } = useChartStore();
  
  const timeframes: { value: Timeframe; label: string; description: string }[] = [
    { value: '1m', label: '1M', description: '1 minute' },
    { value: '5m', label: '5M', description: '5 minutes' },
    { value: '15m', label: '15M', description: '15 minutes' },
    { value: '30m', label: '30M', description: '30 minutes' },
    { value: '1h', label: '1H', description: '1 hour' },
    { value: '4h', label: '4H', description: '4 hours' },
    { value: '1d', label: '1D', description: '1 day' },
  ];
  
  const handleTimeframeChange = (selectedTimeframe: Timeframe) => {
    setTimeframe(selectedTimeframe);
  };

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-card-foreground">
        Timeframe
      </label>
      
      <div className="flex flex-wrap gap-2">
        {timeframes.map((tf) => (
          <button
            key={tf.value}
            onClick={() => handleTimeframeChange(tf.value)}
            className={`px-3 py-2 text-sm font-medium rounded-md transition-colors ${
              timeframe === tf.value
                ? 'bg-primary text-primary-foreground'
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
            title={tf.description}
          >
            {tf.label}
          </button>
        ))}
      </div>
      
      <p className="text-xs text-muted-foreground">
        Current: {timeframes.find(tf => tf.value === timeframe)?.description || timeframe}
      </p>
    </div>
  );
};

export default TimeframeSelector;
