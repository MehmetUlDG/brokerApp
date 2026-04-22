import React, { useRef } from 'react';
import {
  motion,
  useMotionValue,
  useSpring,
  useTransform,
  MotionValue,
} from 'framer-motion';

// ─── Utility: layer depth transform ─────────────────────────────────────────
function useDepthTransform(
  smoothX: MotionValue<number>,
  smoothY: MotionValue<number>,
  depth: number
) {
  const x = useTransform(smoothX, [-0.5, 0.5], [-depth * 22, depth * 22]);
  const y = useTransform(smoothY, [-0.5, 0.5], [-depth * 22, depth * 22]);
  return { x, y };
}

// ─── BTC Price Card ──────────────────────────────────────────────────────────
const PriceCard: React.FC<{ x: MotionValue<number>; y: MotionValue<number> }> = ({ x, y }) => (
  <motion.div
    style={{ x, y }}
    whileHover={{ scale: 1.06, y: -6 }}
    className="absolute top-4 left-6 z-20"
    transition={{ type: 'spring', stiffness: 300, damping: 25 }}
  >
    <div className="glass-card rounded-2xl px-5 py-4 min-w-[170px]"
      style={{
        background: 'rgba(255,255,255,0.05)',
        backdropFilter: 'blur(20px)',
        border: '1px solid rgba(141,198,63,0.25)',
        boxShadow: '0 8px 32px rgba(0,0,0,0.4), 0 0 16px rgba(141,198,63,0.08)',
      }}
    >
      <div className="flex items-center gap-2 mb-2">
        <div className="w-6 h-6 rounded-full bg-amber-400 flex items-center justify-center text-xs font-bold text-black">₿</div>
        <span className="text-xs font-semibold text-white/60 uppercase tracking-widest">BTC/USDT</span>
      </div>
      <div className="text-2xl font-extrabold text-white tracking-tight">$67,420</div>
      <div className="flex items-center gap-1 mt-1">
        <svg className="w-3 h-3 text-[#8DC63F]" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clipRule="evenodd"/>
        </svg>
        <span className="text-xs font-bold text-[#8DC63F]">+3.42%</span>
        <span className="text-xs text-white/30 ml-1">24h</span>
      </div>
    </div>
  </motion.div>
);

// ─── Chart Card ──────────────────────────────────────────────────────────────
const ChartCard: React.FC<{ x: MotionValue<number>; y: MotionValue<number> }> = ({ x, y }) => (
  <motion.div
    style={{ x, y }}
    whileHover={{ scale: 1.04, y: -4 }}
    className="absolute bottom-8 right-4 z-20"
    transition={{ type: 'spring', stiffness: 280, damping: 24 }}
  >
    <div className="rounded-2xl px-5 py-4 min-w-[200px]"
      style={{
        background: 'rgba(255,255,255,0.04)',
        backdropFilter: 'blur(20px)',
        border: '1px solid rgba(255,255,255,0.10)',
        boxShadow: '0 8px 32px rgba(0,0,0,0.5)',
      }}
    >
      <div className="flex justify-between items-center mb-3">
        <span className="text-xs font-semibold text-white/50 uppercase tracking-widest">Price Chart</span>
        <span className="text-xs text-[#8DC63F] font-bold">7D</span>
      </div>
      <svg viewBox="0 0 160 60" className="w-full h-12 overflow-visible">
        <defs>
          <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#8DC63F" stopOpacity="0.4"/>
            <stop offset="100%" stopColor="#8DC63F" stopOpacity="0"/>
          </linearGradient>
        </defs>
        {/* Area fill */}
        <path
          d="M0,55 L0,42 C15,40 25,50 40,35 C55,20 65,30 80,15 C95,5 110,20 130,12 C145,6 155,8 160,5 L160,55 Z"
          fill="url(#chartGrad)"
        />
        {/* Animated line */}
        <path
          d="M0,42 C15,40 25,50 40,35 C55,20 65,30 80,15 C95,5 110,20 130,12 C145,6 155,8 160,5"
          fill="none"
          stroke="#8DC63F"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeDasharray="300"
          strokeDashoffset="0"
          style={{ animation: 'drawLine 2.5s ease forwards' }}
        />
        {/* End dot */}
        <circle cx="160" cy="5" r="4" fill="#8DC63F">
          <animate attributeName="r" values="4;6;4" dur="1.5s" repeatCount="indefinite"/>
          <animate attributeName="opacity" values="1;0.5;1" dur="1.5s" repeatCount="indefinite"/>
        </circle>
      </svg>
    </div>
  </motion.div>
);

// ─── Wallet Card ─────────────────────────────────────────────────────────────
const WalletCard: React.FC<{ x: MotionValue<number>; y: MotionValue<number> }> = ({ x, y }) => (
  <motion.div
    style={{ x, y }}
    whileHover={{ scale: 1.05, y: -6 }}
    className="absolute top-1/2 right-0 -translate-y-1/2 z-20"
    transition={{ type: 'spring', stiffness: 260, damping: 22 }}
  >
    <div className="rounded-2xl px-5 py-4 min-w-[155px]"
      style={{
        background: 'linear-gradient(135deg, rgba(141,198,63,0.15), rgba(141,198,63,0.04))',
        backdropFilter: 'blur(20px)',
        border: '1px solid rgba(141,198,63,0.3)',
        boxShadow: '0 8px 32px rgba(0,0,0,0.4), 0 0 24px rgba(141,198,63,0.12)',
      }}
    >
      <div className="flex items-center gap-2 mb-3">
        <div className="w-7 h-7 rounded-lg bg-[#8DC63F]/20 flex items-center justify-center">
          <svg className="w-4 h-4 text-[#8DC63F]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"/>
          </svg>
        </div>
        <span className="text-xs font-semibold text-white/60 uppercase tracking-widest">Wallet</span>
      </div>
      <div className="text-xl font-extrabold text-white">$12,840</div>
      <div className="text-xs text-[#8DC63F] font-semibold mt-1">Portfolio Value</div>
    </div>
  </motion.div>
);

// ─── ETH mini card ───────────────────────────────────────────────────────────
const EthCard: React.FC<{ x: MotionValue<number>; y: MotionValue<number> }> = ({ x, y }) => (
  <motion.div
    style={{ x, y }}
    whileHover={{ scale: 1.08 }}
    className="absolute top-8 right-14 z-20"
    transition={{ type: 'spring', stiffness: 320, damping: 26 }}
  >
    <div className="rounded-xl px-4 py-3"
      style={{
        background: 'rgba(255,255,255,0.05)',
        backdropFilter: 'blur(16px)',
        border: '1px solid rgba(255,255,255,0.08)',
        boxShadow: '0 4px 16px rgba(0,0,0,0.4)',
      }}
    >
      <div className="flex items-center gap-2">
        <div className="w-5 h-5 rounded-full bg-indigo-400/80 flex items-center justify-center text-[9px] font-bold text-white">Ξ</div>
        <span className="text-sm font-bold text-white">$3,280</span>
        <span className="text-xs font-bold text-[#8DC63F]">+1.8%</span>
      </div>
    </div>
  </motion.div>
);

// ─── Floating stats pill ─────────────────────────────────────────────────────
const StatPill: React.FC<{ x: MotionValue<number>; y: MotionValue<number> }> = ({ x, y }) => (
  <motion.div
    style={{ x, y }}
    className="absolute bottom-14 left-4 z-20"
    transition={{ type: 'spring', stiffness: 200, damping: 20 }}
  >
    <div className="rounded-full px-4 py-2 flex items-center gap-2"
      style={{
        background: 'rgba(141,198,63,0.12)',
        backdropFilter: 'blur(12px)',
        border: '1px solid rgba(141,198,63,0.2)',
      }}
    >
      <div className="w-2 h-2 rounded-full bg-[#8DC63F] animate-pulse"/>
      <span className="text-xs font-bold text-[#8DC63F]">LIVE</span>
      <span className="text-xs text-white/50">Markets Open</span>
    </div>
  </motion.div>
);

// ─── Central Glow Orb ────────────────────────────────────────────────────────
const CenterOrb: React.FC = () => (
  <div className="absolute inset-0 flex items-center justify-center pointer-events-none z-10">
    <div style={{
      width: '280px',
      height: '280px',
      borderRadius: '50%',
      background: 'radial-gradient(circle, rgba(141,198,63,0.12) 0%, transparent 70%)',
      filter: 'blur(20px)',
    }}/>
  </div>
);

// ─── Platform Rings ──────────────────────────────────────────────────────────
const PlatformRings: React.FC = () => (
  <div className="absolute inset-0 flex items-end justify-center pointer-events-none z-0 pb-8">
    <div style={{
      width: '360px',
      height: '60px',
      borderRadius: '50%',
      border: '1px solid rgba(141,198,63,0.15)',
      background: 'transparent',
    }}/>
  </div>
);

// ─── Grid Background ─────────────────────────────────────────────────────────
const GridBg: React.FC = () => (
  <svg className="absolute inset-0 w-full h-full opacity-10" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <pattern id="heroGrid" width="40" height="40" patternUnits="userSpaceOnUse">
        <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(141,198,63,0.5)" strokeWidth="0.5"/>
      </pattern>
    </defs>
    <rect width="100%" height="100%" fill="url(#heroGrid)"/>
  </svg>
);

// ─── Main HeroAnimation Component ────────────────────────────────────────────
export const HeroAnimation: React.FC = () => {
  const containerRef = useRef<HTMLDivElement>(null);

  // Check for touch/mobile — disable tilt
  const isTouch = typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches;

  // Motion values
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);
  const glowX = useMotionValue(50);
  const glowY = useMotionValue(50);

  const springCfg = { stiffness: 150, damping: 20, mass: 0.8 };
  const smoothX = useSpring(mouseX, springCfg);
  const smoothY = useSpring(mouseY, springCfg);

  // Container rotation
  const rotateY = useTransform(smoothX, [-0.5, 0.5], [-12, 12]);
  const rotateX = useTransform(smoothY, [-0.5, 0.5], [10, -10]);

  // Layer depth transforms
  const layer1 = useDepthTransform(smoothX, smoothY, 0.4);  // BTC Card
  const layer2 = useDepthTransform(smoothX, smoothY, 0.7);  // Chart
  const layer3 = useDepthTransform(smoothX, smoothY, 0.55); // Wallet
  const layer4 = useDepthTransform(smoothX, smoothY, 0.85); // ETH mini
  const layer5 = useDepthTransform(smoothX, smoothY, 0.3);  // Stat pill

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (isTouch) return;
    const el = containerRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;
    mouseX.set(x);
    mouseY.set(y);
    glowX.set(((e.clientX - rect.left) / rect.width) * 100);
    glowY.set(((e.clientY - rect.top) / rect.height) * 100);
  };

  const handleMouseLeave = () => {
    mouseX.set(0);
    mouseY.set(0);
    glowX.set(50);
    glowY.set(50);
  };

  return (
    <div
      ref={containerRef}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      style={{ perspective: '1000px' }}
      className="relative w-full max-w-[520px] h-[420px] select-none"
    >
      {/* Rotating scene */}
      <motion.div
        style={isTouch ? {} : { rotateX, rotateY, transformStyle: 'preserve-3d' }}
        className="relative w-full h-full"
        transition={{ type: 'spring' }}
      >
        <GridBg />
        <CenterOrb />
        <PlatformRings />

        {/* Cursor glow that follows mouse */}
        <motion.div
          className="pointer-events-none absolute z-30 rounded-full"
          style={{
            width: 180,
            height: 180,
            background: 'radial-gradient(circle, rgba(141,198,63,0.18) 0%, transparent 70%)',
            filter: 'blur(20px)',
            left: useTransform(glowX, (v) => `calc(${v}% - 90px)`),
            top: useTransform(glowY, (v) => `calc(${v}% - 90px)`),
          }}
        />

        {/* Central isometric-ish platform box */}
        <div className="absolute inset-0 flex items-center justify-center z-10">
          <div style={{
            width: '200px',
            height: '120px',
            background: 'linear-gradient(145deg, rgba(141,198,63,0.06), rgba(141,198,63,0.02))',
            backdropFilter: 'blur(12px)',
            border: '1px solid rgba(141,198,63,0.2)',
            borderRadius: '20px',
            boxShadow: '0 20px 60px rgba(0,0,0,0.5), 0 0 40px rgba(141,198,63,0.08)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}>
            <svg width="80" height="56" viewBox="0 0 80 56" fill="none">
              {/* Candlestick chart */}
              {[
                { x: 8,  h: 20, y: 20, color: '#ef4444' },
                { x: 20, h: 30, y: 10, color: '#8DC63F' },
                { x: 32, h: 15, y: 25, color: '#8DC63F' },
                { x: 44, h: 25, y: 15, color: '#ef4444' },
                { x: 56, h: 35, y: 8,  color: '#8DC63F' },
                { x: 68, h: 28, y: 12, color: '#8DC63F' },
              ].map((bar, i) => (
                <g key={i}>
                  <rect x={bar.x} y={bar.y} width="8" height={bar.h} rx="2" fill={bar.color} opacity="0.85"/>
                  <line x1={bar.x + 4} y1={bar.y - 4} x2={bar.x + 4} y2={bar.y} stroke={bar.color} strokeWidth="1.5" opacity="0.6"/>
                  <line x1={bar.x + 4} y1={bar.y + bar.h} x2={bar.x + 4} y2={bar.y + bar.h + 4} stroke={bar.color} strokeWidth="1.5" opacity="0.6"/>
                </g>
              ))}
            </svg>
          </div>
        </div>

        {/* Depth layers */}
        <PriceCard x={layer1.x} y={layer1.y} />
        <ChartCard x={layer2.x} y={layer2.y} />
        <WalletCard x={layer3.x} y={layer3.y} />
        <EthCard x={layer4.x} y={layer4.y} />
        <StatPill x={layer5.x} y={layer5.y} />
      </motion.div>
    </div>
  );
};
