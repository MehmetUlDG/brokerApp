import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { LogOut, Sun, Moon } from 'lucide-react';
import { useTheme } from '../theme/ThemeContext';

export const Navbar: React.FC = () => {
  const { isAuthenticated, logout, user } = useAuthStore();
  const location = useLocation();
  const { theme, toggleTheme } = useTheme();

  const handleLogout = () => {
    logout();
  };

  const isDark = theme === 'dark';

  return (
    <header className={`sticky top-0 z-50 w-full transition-colors duration-300 ${location.pathname === '/' ? 'bg-transparent' : 'border-b border-border bg-background/95 backdrop-blur'}`}>
      <div className="max-w-7xl mx-auto px-8 h-20 flex items-center justify-between w-full">
        
        {/* Left Section: Logo & Links */}
        <div className="flex items-center gap-12">
          {/* TRADEX Logo */}
          <Link to="/" className="flex items-center gap-3 group">
            <div className="w-6 h-6 flex flex-col justify-between">
               <div className="h-1.5 w-full bg-[#84cc16] rounded-sm group-hover:bg-[#94e01b] transition-colors"></div>
               <div className="h-1.5 w-full bg-[#84cc16] rounded-sm group-hover:bg-[#94e01b] transition-colors"></div>
               <div className="h-1.5 w-full bg-[#84cc16] rounded-sm group-hover:bg-[#94e01b] transition-colors"></div>
            </div>
            <span className="font-bold text-2xl tracking-widest uppercase text-foreground">Tradex</span>
          </Link>
          
          {/* Navigation Links */}
          <nav className="hidden md:flex items-center gap-8">
            <Link to="/" className={`text-sm font-semibold tracking-wider uppercase transition-colors hover:text-primary ${location.pathname === '/' ? 'text-primary' : (isDark ? 'text-slate-400' : 'text-slate-500')}`}>
              HOME
            </Link>
            {isAuthenticated && (
              <>
                <Link to="/trade" className={`text-sm font-semibold tracking-wider uppercase transition-colors hover:text-primary ${location.pathname.startsWith('/trade') ? 'text-primary' : (isDark ? 'text-slate-400' : 'text-slate-500')}`}>
                  TRADE
                </Link>
                <Link to="/wallet" className={`text-sm font-semibold tracking-wider uppercase transition-colors hover:text-primary ${location.pathname.startsWith('/wallet') ? 'text-primary' : (isDark ? 'text-slate-400' : 'text-slate-500')}`}>
                  WALLET
                </Link>
              </>
            )}
          </nav>
        </div>

        {/* Right Section: Theme Toggle & Auth */}
        <div className="flex items-center gap-6">
          <button
            onClick={toggleTheme}
            className="p-2 rounded-full hover:bg-black/5 dark:hover:bg-white/10 transition-colors duration-300 flex items-center justify-center mr-2"
            aria-label="Toggle theme"
          >
            {isDark ? (
              <Sun className="h-5 w-5 text-yellow-400 transition-all duration-300" />
            ) : (
              <Moon className="h-5 w-5 text-slate-700 transition-all duration-300" />
            )}
          </button>

          {isAuthenticated ? (
            <div className="flex items-center gap-4">
              <span className="text-sm font-medium hidden sm:inline-block text-foreground">
                {user?.first_name} {user?.last_name}
              </span>
              <button 
                onClick={handleLogout}
                className="p-2 text-muted-foreground hover:text-destructive transition-colors rounded-full hover:bg-muted"
                title="Logout"
              >
                <LogOut className="h-5 w-5" />
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-8">
              <Link to="/login" className="text-sm font-bold text-foreground hover:text-primary transition-colors">
                Login
              </Link>
              <Link to="/register" className="bg-[#84cc16] hover:bg-[#74b814] text-[#0a0a0a] px-8 py-2.5 rounded-full font-bold text-sm transition-colors shadow-sm">
                Register
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};

