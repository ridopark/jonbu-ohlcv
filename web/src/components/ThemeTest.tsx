import React from 'react';
import { useThemeStore } from '../stores/themeStore';
import ThemeToggle from '../components/ui/ThemeToggle';

const ThemeTest: React.FC = () => {
  const { theme, resolvedTheme } = useThemeStore();

  return (
    <div className="p-8 space-y-6">
      <div className="bg-card p-6 rounded-lg border border-border">
        <h1 className="text-2xl font-bold text-card-foreground mb-4">Dark Mode Test</h1>
        
        <div className="space-y-4">
          <div>
            <strong>Current Theme:</strong> {theme}
          </div>
          <div>
            <strong>Resolved Theme:</strong> {resolvedTheme}
          </div>
          
          <ThemeToggle />
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-6">
            <div className="bg-primary p-4 rounded-lg">
              <h3 className="text-primary-foreground font-semibold">Primary Card</h3>
              <p className="text-primary-foreground/80">This should work in both themes</p>
            </div>
            
            <div className="bg-muted p-4 rounded-lg">
              <h3 className="text-card-foreground font-semibold">Muted Card</h3>
              <p className="text-muted-foreground">Secondary content</p>
            </div>
            
            <div className="bg-accent p-4 rounded-lg">
              <h3 className="text-card-foreground font-semibold">Accent Card</h3>
              <p className="text-muted-foreground">Hover state background</p>
            </div>
            
            <div className="bg-destructive/10 border border-destructive/20 p-4 rounded-lg">
              <h3 className="text-destructive font-semibold">Error State</h3>
              <p className="text-muted-foreground">Destructive styling</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ThemeTest;
