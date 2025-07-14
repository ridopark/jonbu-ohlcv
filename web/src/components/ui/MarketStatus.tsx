import React from 'react';

interface MarketStatusProps {
  className?: string;
}

const MarketStatus: React.FC<MarketStatusProps> = ({ className = '' }) => {
  const [status, setStatus] = React.useState<'open' | 'closed' | 'pre-market' | 'after-hours'>('closed');
  const [marketTime, setMarketTime] = React.useState<string>('');

  React.useEffect(() => {
    const updateMarketStatus = () => {
      const now = new Date();
      const eastern = new Date(now.toLocaleString("en-US", { timeZone: "America/New_York" }));
      const hours = eastern.getHours();
      const minutes = eastern.getMinutes();
      const day = eastern.getDay(); // 0 = Sunday, 6 = Saturday
      
      // Update time display
      setMarketTime(eastern.toLocaleTimeString('en-US', {
        timeZone: 'America/New_York',
        hour12: true,
        hour: 'numeric',
        minute: '2-digit'
      }));
      
      // Weekend
      if (day === 0 || day === 6) {
        setStatus('closed');
        return;
      }
      
      // Market hours: 9:30 AM - 4:00 PM ET
      const marketOpen = 9 * 60 + 30; // 9:30 AM in minutes
      const marketClose = 16 * 60; // 4:00 PM in minutes
      const currentMinutes = hours * 60 + minutes;
      
      if (currentMinutes >= marketOpen && currentMinutes < marketClose) {
        setStatus('open');
      } else if (currentMinutes >= 4 * 60 && currentMinutes < marketOpen) {
        setStatus('pre-market');
      } else if (currentMinutes >= marketClose && currentMinutes < 20 * 60) {
        setStatus('after-hours');
      } else {
        setStatus('closed');
      }
    };
    
    // Update immediately
    updateMarketStatus();
    
    // Update every minute
    const interval = setInterval(updateMarketStatus, 60000);
    
    return () => clearInterval(interval);
  }, []);

  const getStatusConfig = () => {
    switch (status) {
      case 'open':
        return {
          color: 'text-green-600 dark:text-green-400',
          bg: 'bg-green-100 dark:bg-green-900/20',
          dot: 'bg-green-500',
          text: 'Market Open'
        };
      case 'pre-market':
        return {
          color: 'text-yellow-600 dark:text-yellow-400',
          bg: 'bg-yellow-100 dark:bg-yellow-900/20',
          dot: 'bg-yellow-500',
          text: 'Pre-Market'
        };
      case 'after-hours':
        return {
          color: 'text-blue-600 dark:text-blue-400',
          bg: 'bg-blue-100 dark:bg-blue-900/20',
          dot: 'bg-blue-500',
          text: 'After Hours'
        };
      case 'closed':
      default:
        return {
          color: 'text-red-600 dark:text-red-400',
          bg: 'bg-red-100 dark:bg-red-900/20',
          dot: 'bg-red-500',
          text: 'Market Closed'
        };
    }
  };

  const config = getStatusConfig();

  return (
    <div className={`flex items-center gap-2 px-3 py-1 rounded-full ${config.bg} ${className}`}>
      <div className={`w-2 h-2 rounded-full ${config.dot} ${status === 'open' ? 'animate-pulse' : ''}`} />
      <div className="flex flex-col sm:flex-row sm:items-center sm:gap-2">
        <span className={`text-sm font-medium ${config.color}`}>
          {config.text}
        </span>
        <span className="text-xs text-muted-foreground">
          {marketTime} ET
        </span>
      </div>
    </div>
  );
};

export default MarketStatus;
