import React from 'react';

const Symbols: React.FC = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-card-foreground">Symbols</h1>
        <p className="text-muted-foreground">
          Manage tracked symbols and their configurations
        </p>
      </div>
      
      <div className="bg-card rounded-lg border border-border p-8 text-center">
        <div className="text-4xl mb-4">🏷️</div>
        <h3 className="text-lg font-medium text-card-foreground mb-2">
          Symbol Management
        </h3>
        <p className="text-muted-foreground mb-4">
          Add, remove, and configure symbols for real-time tracking
        </p>
        <button className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors">
          Coming Soon
        </button>
      </div>
    </div>
  );
};

export default Symbols;
