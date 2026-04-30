'use client';

import { useLivePrice } from '@/hooks/useLivePrice';
import { useTradeStore } from '@/stores/tradeStore';
import { useEffect, useState } from 'react';
import { ArrowDownRight, ArrowUpRight } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

export function TickerBar() {
  useLivePrice();
  const livePrice = useTradeStore((state) => state.livePrice);
  const priceHistory = useTradeStore((state) => state.priceHistory);
  
  const [direction, setDirection] = useState<'up' | 'down' | null>(null);

  useEffect(() => {
    if (priceHistory.length >= 2) {
      const current = priceHistory[priceHistory.length - 1].value;
      const previous = priceHistory[priceHistory.length - 2].value;
      if (current > previous) setDirection('up');
      else if (current < previous) setDirection('down');
    }
  }, [priceHistory]);

  const priceNum = parseFloat(livePrice || '0');
  const formattedPrice = priceNum > 0 ? priceNum.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : '...';

  return (
    <div className="flex items-center h-16 px-6 bg-[var(--surface)] border-b border-[var(--border)] gap-8 overflow-x-auto whitespace-nowrap">
      <div className="flex items-center gap-4">
        <h2 className="text-xl font-bold text-[var(--text-primary)]">BTC/USDT</h2>
        <div className={cn(
          "flex items-center text-lg font-mono transition-colors",
          direction === 'up' ? 'text-[var(--success)]' : direction === 'down' ? 'text-[var(--danger)]' : 'text-[var(--text-primary)]'
        )}>
          {formattedPrice}
          {direction === 'up' && <ArrowUpRight className="h-4 w-4 ml-1" />}
          {direction === 'down' && <ArrowDownRight className="h-4 w-4 ml-1" />}
        </div>
      </div>
      
      <div className="hidden sm:flex flex-col">
        <span className="text-xs text-[var(--text-muted)]">24s Değişim</span>
        <span className="text-sm font-medium text-[var(--success)]">+2.45%</span>
      </div>
      <div className="hidden md:flex flex-col">
        <span className="text-xs text-[var(--text-muted)]">24s En Yüksek</span>
        <span className="text-sm font-medium text-[var(--text-primary)]">{(priceNum * 1.02).toLocaleString()}</span>
      </div>
      <div className="hidden md:flex flex-col">
        <span className="text-xs text-[var(--text-muted)]">24s En Düşük</span>
        <span className="text-sm font-medium text-[var(--text-primary)]">{(priceNum * 0.98).toLocaleString()}</span>
      </div>
      <div className="hidden lg:flex flex-col">
        <span className="text-xs text-[var(--text-muted)]">24s Hacim(BTC)</span>
        <span className="text-sm font-medium text-[var(--text-primary)]">24,531.42</span>
      </div>
    </div>
  );
}
