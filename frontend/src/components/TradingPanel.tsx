import React, { useState, useEffect } from 'react';
import { motion, animate, AnimatePresence } from 'framer-motion';
import toast from 'react-hot-toast';
import { useWebSocket } from '../hooks/useWebSocket';
import { apiService, type OrderPayload } from '../services/api';
import { useTradingStore } from '../store/tradingStore';

// Animated Number Component
const AnimatedNumber: React.FC<{ value: number; prefix?: string; decimals?: number }> = ({ value, prefix = '', decimals = 2 }) => {
  const [displayValue, setDisplayValue] = useState(value);

  useEffect(() => {
    const controls = animate(displayValue, value, {
      duration: 0.6,
      ease: "easeOut",
      onUpdate: (v) => setDisplayValue(v),
    });
    return controls.stop;
  }, [value]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <span>
      {prefix}{displayValue.toLocaleString('en-US', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })}
    </span>
  );
};

export const TradingPanel: React.FC = () => {
  const wsUrl = import.meta.env.VITE_WS_URL || (import.meta.env as any).REACT_APP_WS_URL || 'ws://localhost:8080/ws/btc';
  const { status: wsStatus } = useWebSocket(wsUrl);
  const btcData = useTradingStore((state) => state.btc);

  // Form State
  const [side, setSide] = useState<'buy' | 'sell'>('buy');
  const [orderType, setOrderType] = useState<'market' | 'limit'>('market');
  const [price, setPrice] = useState<string>('');
  const [quantity, setQuantity] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showConfirmModal, setShowConfirmModal] = useState(false);

  // Derived data
  const currentPrice = btcData?.price || 64230.50; // Fallback to mock if no data
  const totalValue = parseFloat(quantity || '0') * (orderType === 'market' ? currentPrice : parseFloat(price || '0'));

  const handlePlaceOrderClick = () => {
    const qty = parseFloat(quantity);
    if (isNaN(qty) || qty < 0.0001) {
      toast.error('Minimum quantity is 0.0001 BTC');
      return;
    }
    if (orderType === 'limit') {
      const p = parseFloat(price);
      if (isNaN(p) || p <= 0) {
        toast.error('Enter a valid limit price');
        return;
      }
    }
    setShowConfirmModal(true);
  };

  const executeOrder = async () => {
    setIsSubmitting(true);
    try {
      const payload: OrderPayload = {
        symbol: 'BTC/USDT',
        side,
        type: orderType,
        quantity: parseFloat(quantity),
      };
      if (orderType === 'limit') {
        payload.price = parseFloat(price);
      }

      const response = await apiService.placeOrder(payload);
      toast.success(`Order Placed Successfully! ID: ${response?.id || 'Tx-' + Math.floor(Math.random() * 10000)}`);

      // Reset form
      setQuantity('');
      if (orderType === 'limit') setPrice('');
      setShowConfirmModal(false);
    } catch (error) {
      toast.error('Failed to place order. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] w-full bg-[#111827] text-white p-6 relative overflow-hidden font-sans flex items-center justify-center">
      {/* Background Grid Pattern */}
      <div
        className="absolute inset-0 z-0 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage: `linear-gradient(rgba(255,255,255,1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,1) 1px, transparent 1px)`,
          backgroundSize: '30px 30px'
        }}
      />

      <div className="max-w-6xl w-full z-10 relative">

        {/* Top Header Row */}
        <div className="flex flex-col md:flex-row justify-between items-start md:items-end mb-12 gap-6">
          <div className="max-w-xl">
            <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight mb-4">Real Time Trading</h1>
            <p className="text-slate-400 text-lg">Trade top cryptocurrencies on the broker platform.</p>
          </div>

          {/* Ticker Badge */}
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-[#1F2937] border border-slate-700/50 rounded-2xl px-5 py-3 flex items-center gap-4 shadow-xl"
          >
            <div className="flex flex-col items-center">
              <div className="flex items-center gap-2 mb-1">
                {wsStatus === 'connected' ? (
                  <span className="relative flex h-3 w-3">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#84CC16] opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-[#84CC16]"></span>
                  </span>
                ) : (
                  <span className="relative flex h-3 w-3">
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-yellow-500"></span>
                  </span>
                )}
                <span className="text-xs font-semibold text-slate-400">
                  {wsStatus === 'connected' ? 'LIVE' : 'RECONNECTING'}
                </span>
              </div>
              <span className="text-xs font-medium text-slate-500">BTC/USDT</span>
            </div>
            <div className="text-2xl font-bold tracking-tight">
              <AnimatedNumber value={currentPrice} decimals={2} />
            </div>
          </motion.div>
        </div>

        {/* Main Panels */}
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">

          {/* Left Hero Card */}
          <div className="lg:col-span-3">
            <div className="bg-[#1C2128] border border-[#30363D] rounded-[2rem] p-10 h-full shadow-2xl flex flex-col justify-center items-center text-center relative overflow-hidden group">
              <div className="absolute top-0 left-0 w-full h-full bg-gradient-to-b from-[#84CC16]/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

              <div className="bg-[#84CC16] text-[#111827] text-xs font-bold px-3 py-1 rounded-full mb-8 z-10">
                FEATURED
              </div>

              <h2 className="text-4xl md:text-5xl font-bold mb-6 z-10 leading-tight">
                Cryptocurrency Trading<br />App
              </h2>

              <p className="text-slate-400 text-lg max-w-md mx-auto z-10">
                Advanced charting and real time order execution within milliseconds.
              </p>

              {/* Live OHLCV Data Overlay (Hidden on small screens, shown nicely) */}
              <div className="mt-12 grid grid-cols-2 md:grid-cols-4 gap-4 w-full z-10">
                <div className="bg-black/20 rounded-xl p-4 border border-white/5">
                  <p className="text-xs text-slate-500 mb-1">24h Change</p>
                  <p className={`text-lg font-bold ${btcData && btcData.change24h < 0 ? 'text-red-500' : 'text-[#84CC16]'}`}>
                    {btcData ? `${btcData.change24h > 0 ? '+' : ''}${btcData.change24h}%` : '+2.45%'}
                  </p>
                </div>
                <div className="bg-black/20 rounded-xl p-4 border border-white/5">
                  <p className="text-xs text-slate-500 mb-1">24h High</p>
                  <p className="text-lg font-bold">
                    <AnimatedNumber value={btcData?.high24h || 65420.00} />
                  </p>
                </div>
                <div className="bg-black/20 rounded-xl p-4 border border-white/5">
                  <p className="text-xs text-slate-500 mb-1">24h Low</p>
                  <p className="text-lg font-bold">
                    <AnimatedNumber value={btcData?.low24h || 62100.00} />
                  </p>
                </div>
                <div className="bg-black/20 rounded-xl p-4 border border-white/5">
                  <p className="text-xs text-slate-500 mb-1">Volume</p>
                  <p className="text-lg font-bold">
                    <AnimatedNumber value={btcData?.volume || 18432.5} decimals={1} />
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Right Order Form */}
          <div className="lg:col-span-2">
            <div className="bg-[#1C2128] border border-[#30363D] rounded-3xl p-8 shadow-2xl relative">
              {/* Subtle glow behind the form */}
              <div className="absolute -inset-4 bg-[#84CC16]/5 blur-3xl rounded-full pointer-events-none" />

              <div className="relative z-10">
                <div className="bg-[#84CC16] text-[#111827] text-xs font-bold px-3 py-1 rounded-full inline-block mb-6">
                  ORDER ENTRY
                </div>

                <h3 className="text-3xl font-bold mb-2">Place Order</h3>
                <p className="text-slate-400 text-sm mb-8">Execute your trading strategy.</p>

                {/* Symbol & Side */}
                <div className="flex gap-4 mb-6">
                  <div className="flex-1">
                    <label className="text-xs text-slate-500 font-semibold uppercase mb-2 block">Symbol</label>
                    <div className="bg-[#111827] border border-[#30363D] rounded-xl p-3 text-sm font-medium flex justify-between items-center opacity-80 cursor-not-allowed">
                      BTC/USDT
                      <svg className="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7"></path></svg>
                    </div>
                  </div>

                  <div className="flex-[1.2]">
                    <label className="text-xs text-slate-500 font-semibold uppercase mb-2 block">Side</label>
                    <div className="flex bg-[#111827] rounded-xl p-1 border border-[#30363D]">
                      <button
                        onClick={() => setSide('buy')}
                        className={`flex-1 text-sm font-bold py-2 rounded-lg transition-all ${side === 'buy' ? 'bg-[#84CC16] text-[#111827] shadow-lg' : 'text-slate-400 hover:text-white'}`}
                      >
                        BUY
                      </button>
                      <button
                        onClick={() => setSide('sell')}
                        className={`flex-1 text-sm font-bold py-2 rounded-lg transition-all ${side === 'sell' ? 'bg-red-500 text-white shadow-lg' : 'text-slate-400 hover:text-white'}`}
                      >
                        SELL
                      </button>
                    </div>
                  </div>
                </div>

                {/* Order Type */}
                <div className="mb-6">
                  <label className="text-xs text-slate-500 font-semibold uppercase mb-2 block">Order Type</label>
                  <div className="flex gap-6">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="radio"
                        name="type"
                        checked={orderType === 'market'}
                        onChange={() => setOrderType('market')}
                        className="w-4 h-4 text-[#84CC16] bg-[#111827] border-[#30363D] focus:ring-[#84CC16]"
                      />
                      <span className="text-sm font-medium">Market</span>
                    </label>
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="radio"
                        name="type"
                        checked={orderType === 'limit'}
                        onChange={() => setOrderType('limit')}
                        className="w-4 h-4 text-[#84CC16] bg-[#111827] border-[#30363D] focus:ring-[#84CC16]"
                      />
                      <span className="text-sm font-medium">Limit</span>
                    </label>
                  </div>
                </div>

                {/* Price */}
                <div className="mb-6">
                  <div className="flex justify-between mb-2">
                    <label className="text-xs text-slate-500 font-semibold uppercase block">Price (USDT)</label>
                    {orderType === 'market' && <span className="text-xs font-semibold text-[#84CC16] italic">Market Price</span>}
                  </div>
                  <input
                    type={orderType === 'market' ? 'text' : 'number'}
                    disabled={orderType === 'market'}
                    value={orderType === 'market' ? 'Market Execution' : price}
                    onChange={(e) => setPrice(e.target.value)}
                    placeholder="Enter limit price"
                    className={`w-full bg-[#111827] border ${orderType === 'limit' ? 'border-[#84CC16]/50 focus:border-[#84CC16] text-white' : 'border-[#30363D] text-slate-500'} rounded-xl p-4 text-sm outline-none transition-colors disabled:opacity-70`}
                  />
                </div>

                {/* Quantity */}
                <div className="mb-8">
                  <label className="text-xs text-slate-500 font-semibold uppercase mb-2 block">Quantity (COIN)</label>
                  <div className="relative">
                    <input
                      type="number"
                      step="0.0001"
                      min="0.0001"
                      value={quantity}
                      onChange={(e) => setQuantity(e.target.value)}
                      placeholder="0.00"
                      className="w-full bg-[#111827] border border-[#30363D] focus:border-[#84CC16] rounded-xl p-4 text-sm text-white outline-none transition-colors"
                    />
                    <div className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 text-xs font-medium">
                      ~ ${totalValue > 0 ? totalValue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : '0.00'}
                    </div>
                  </div>
                </div>

                {/* Total & Submit */}
                <div className="flex items-center justify-between mb-6">
                  <span className="text-sm text-slate-400 font-medium">Total Volume</span>
                  <span className="text-xl font-bold">
                    ${totalValue > 0 ? totalValue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : '0.00'}
                  </span>
                </div>

                <button
                  onClick={handlePlaceOrderClick}
                  disabled={isSubmitting}
                  className={`w-full py-4 rounded-xl font-bold text-lg transition-all transform active:scale-[0.98] flex items-center justify-center gap-2 ${side === 'buy'
                      ? 'bg-[#84CC16] hover:bg-[#65A30D] text-[#111827]'
                      : 'bg-red-500 hover:bg-red-600 text-white'
                    }`}
                >
                  {isSubmitting ? (
                    <div className="w-6 h-6 border-2 border-current border-t-transparent rounded-full animate-spin" />
                  ) : (
                    `PLACE ${side.toUpperCase()} ORDER`
                  )}
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>

      {/* Confirmation Modal */}
      <AnimatePresence>
        {showConfirmModal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
          >
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              className="bg-[#1C2128] border border-[#30363D] rounded-2xl p-6 max-w-sm w-full shadow-2xl"
            >
              <h3 className="text-xl font-bold mb-4">Confirm Order</h3>
              <div className="space-y-3 mb-6 text-sm bg-[#111827] p-4 rounded-xl">
                <div className="flex justify-between"><span className="text-slate-400">Pair</span> <span className="font-bold">BTC/USDT</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Side</span> <span className={`font-bold ${side === 'buy' ? 'text-[#84CC16]' : 'text-red-500'}`}>{side.toUpperCase()}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Type</span> <span className="font-bold capitalize">{orderType}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Quantity</span> <span className="font-bold">{quantity} BTC</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Est. Total</span> <span className="font-bold">${totalValue.toLocaleString(undefined, { minimumFractionDigits: 2 })}</span></div>
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowConfirmModal(false)}
                  className="flex-1 py-3 bg-[#111827] hover:bg-[#30363D] text-white rounded-lg font-semibold transition-colors border border-[#30363D]"
                >
                  Cancel
                </button>
                <button
                  onClick={executeOrder}
                  className={`flex-1 py-3 rounded-lg font-bold transition-colors ${side === 'buy' ? 'bg-[#84CC16] hover:bg-[#65A30D] text-[#111827]' : 'bg-red-500 hover:bg-red-600 text-white'}`}
                >
                  Confirm
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
