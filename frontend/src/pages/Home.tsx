import React, { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { TradingPanel } from '../components/TradingPanel';
import { WalletSection } from '../components/WalletSection';
import { useTradingStore } from '../store/tradingStore';
import { useTheme } from '../theme/ThemeContext';


export const Home: React.FC = () => {
  const { fetchBalance, fetchTransactions } = useTradingStore();
  const { theme } = useTheme();
  const isDark = theme === 'dark';

  useEffect(() => {
    fetchBalance();
    fetchTransactions();
  }, [fetchBalance, fetchTransactions]);

  return (
    <div
      style={{
        minHeight: '100vh',
        background: '#0B0E14', // Match the dark background of the image
        backgroundImage: 'radial-gradient(ellipse 80% 80% at 50% -20%, rgba(132, 204, 22, 0.05) 0%, transparent 100%)',
        overflow: 'hidden',
      }}
      className="text-white"
    >
      {/* ── Hero Section (New Design) ──────────────────────────────────── */}
      <div className="max-w-7xl mx-auto px-8 py-20 flex flex-col lg:flex-row items-center justify-between relative min-h-[calc(100vh-80px)]">
        <div className="max-w-2xl z-10">
          <h1 className="text-6xl lg:text-7xl font-bold leading-[1.1] mb-6 tracking-tight">
            Welcome<br/>to Tradex
          </h1>
          <p className={`text-lg mb-10 max-w-md ${isDark ? 'text-slate-400' : 'text-slate-500'}`}>
            Trade securely and manage your wallet with our next-generation platform.
          </p>
          <div className="flex items-center gap-4">
            <Link to="/login" className="bg-[#84cc16] text-[#0a0a0a] px-10 py-3.5 rounded-full font-bold text-sm shadow-[0_0_30px_rgba(132,204,22,0.3)] hover:bg-[#94e01b] transition-colors text-center">
              LOGIN
            </Link>
            <Link to="/register" className={`border px-10 py-3.5 rounded-full font-bold text-sm transition-colors text-center ${isDark ? 'border-white/20 hover:bg-white/5' : 'border-slate-300 hover:bg-slate-50'}`}>
              REGISTER
            </Link>
          </div>
        </div>
        <div className="w-full lg:w-1/2 mt-16 lg:mt-0 relative flex justify-center lg:justify-end">
           <div className="relative w-full max-w-lg aspect-square">
             <div className="absolute inset-0 bg-[#84cc16]/20 blur-[100px] rounded-full"></div>
             {/* 3D Glassmorphic Cube Illustration Replica */}
             <div className="relative z-10 w-full h-full flex flex-wrap items-center justify-center gap-4 transform rotate-[-15deg] skew-x-[10deg] scale-90 pointer-events-none">
               <div className={`w-32 h-32 rounded-xl border-4 ${isDark ? 'bg-white/5 border-white/10' : 'bg-black/5 border-black/10'} shadow-[0_0_30px_rgba(132,204,22,0.4)] backdrop-blur-md flex items-center justify-center`}><div className="w-16 h-16 bg-[#84cc16] rounded-lg animate-pulse"></div></div>
               <div className={`w-24 h-24 rounded-xl border-4 ${isDark ? 'bg-white/5 border-white/10' : 'bg-black/5 border-black/10'} shadow-[0_0_20px_rgba(132,204,22,0.3)] backdrop-blur-md mt-20 flex items-center justify-center`}><div className="w-12 h-12 bg-[#84cc16] rounded-lg"></div></div>
               <div className={`w-40 h-40 rounded-xl border-4 ${isDark ? 'bg-white/5 border-white/10' : 'bg-black/5 border-black/10'} shadow-[0_0_40px_rgba(132,204,22,0.5)] backdrop-blur-md -mt-20 flex items-center justify-center`}><div className="w-20 h-20 bg-[#84cc16] rounded-lg"></div></div>
             </div>
           </div>
        </div>
      </div>

      {/* ── Additional Sections requested previously ─────────────────── */}
      <div className="w-full flex flex-col gap-12 pb-24 relative z-10"> </div>
    </div>
  );
};
