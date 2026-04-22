import React from 'react';
import { Outlet } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { Navbar } from './components/Navbar';
import { ThemeProvider } from './theme/ThemeContext';
import { useTradingStore } from './store/tradingStore';

const AppLayout: React.FC = () => {
  const { networkError } = useTradingStore();

  return (
    <ThemeProvider>
      <div className="min-h-screen bg-background text-foreground flex flex-col font-sans transition-colors duration-300">
        <Toaster 
          position="bottom-center"
          toastOptions={{
            style: {
              background: '#1C1C1C',
              color: '#fff',
              border: '1px solid #2D2D2D'
            },
          }}
        />

        {/* Network Error Banner */}
        {networkError && (
          <div className="w-full bg-red-500 text-white text-center py-2 text-sm font-bold shadow-md z-50">
            Connection lost. Please check your internet connection.
          </div>
        )}
        
        {/* Top Navbar */}
        <Navbar />

        {/* Main Content Area */}
        <main className="flex-1">
          <Outlet />
        </main>
        
        {/* Footer */}
        <footer className="border-t border-border py-6 md:py-0 bg-[#000000]">
          <div className="container mx-auto px-4 flex flex-col md:h-16 items-center justify-between md:flex-row">
            <p className="text-sm text-muted-foreground">
              &copy; 2026 Tradex Demo. All rights reserved.
            </p>
          </div>
        </footer>
      </div>
    </ThemeProvider>
  );
};

export default AppLayout;
