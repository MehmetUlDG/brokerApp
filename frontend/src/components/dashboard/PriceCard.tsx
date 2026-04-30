'use client';

import { Card } from '@/components/ui/Card';
import { useTradeStore } from '@/stores/tradeStore';
import { useLivePrice } from '@/hooks/useLivePrice';
import { useEffect, useState } from 'react';
import { ArrowDownRight, ArrowUpRight } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

export function PriceCard() {
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
    <Card className="p-6">
      <div className="flex items-center gap-3 mb-2">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#F7931A]/10">
          <span className="text-lg font-bold text-[#F7931A]">₿</span>
        </div>
        <div>
          <h3 className="font-bold text-[var(--text-primary)]">BTC/USDT</h3>
          <p className="text-sm text-[var(--text-secondary)]">Bitcoin</p>
        </div>
      </div>
      
      <div className="mt-4 flex items-end gap-2">
        <div className={cn(
          "text-3xl font-mono transition-colors duration-300",
          direction === 'up' ? 'text-[var(--success)]' : direction === 'down' ? 'text-[var(--danger)]' : 'text-[var(--text-primary)]'
        )}>
          ${formattedPrice}
        </div>
        {direction === 'up' && <ArrowUpRight className="h-6 w-6 text-[var(--success)] mb-1" />}
        {direction === 'down' && <ArrowDownRight className="h-6 w-6 text-[var(--danger)] mb-1" />}
      </div>
    </Card>
  );
}
