import React from 'react';
import { useChartStore } from '../../stores/chartStore';

const SymbolSelector: React.FC = () => {
  const { symbol, setSymbol } = useChartStore();
  
  const popularSymbols = [
    'AAPL', 'GOOGL', 'MSFT', 'AMZN', 'TSLA', 
    'META', 'NVDA', 'NFLX', 'AMD', 'INTC'
  ];
  
  const handleSymbolChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setSymbol(event.target.value);
  };
  
  const handleCustomSymbol = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value.toUpperCase();
    if (/^[A-Z]*$/.test(value)) { // Only allow uppercase letters
      setSymbol(value);
    }
  };

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-card-foreground">
        Symbol
      </label>
      
      <div className="flex gap-2">
        {/* Popular symbols dropdown */}
        <select
          value={popularSymbols.includes(symbol) ? symbol : ''}
          onChange={handleSymbolChange}
          className="flex-1 px-3 py-2 bg-background border border-input rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
        >
          <option value="">Select symbol...</option>
          {popularSymbols.map((sym) => (
            <option key={sym} value={sym}>
              {sym}
            </option>
          ))}
        </select>
        
        {/* Custom symbol input */}
        <input
          type="text"
          value={symbol}
          onChange={handleCustomSymbol}
          placeholder="Or type symbol"
          className="flex-1 px-3 py-2 bg-background border border-input rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          maxLength={10}
        />
      </div>
      
      <p className="text-xs text-muted-foreground">
        Select a popular symbol or enter a custom one (e.g., AAPL, GOOGL)
      </p>
    </div>
  );
};

export default SymbolSelector;
