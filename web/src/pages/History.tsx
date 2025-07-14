import React from 'react';

const History: React.FC = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-card-foreground">History</h1>
        <p className="text-muted-foreground">
          View and analyze historical OHLCV data
        </p>
      </div>
      
      <div className="bg-card rounded-lg border border-border p-8 text-center">
        <div className="text-4xl mb-4">📈</div>
        <h3 className="text-lg font-medium text-card-foreground mb-2">
          Historical Data Analysis
        </h3>
        <p className="text-muted-foreground mb-4">
          Explore historical market data with advanced filtering and analysis tools
        </p>
        <button className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors">
          Coming Soon
        </button>
      </div>
    </div>
  );
};

export default History;
