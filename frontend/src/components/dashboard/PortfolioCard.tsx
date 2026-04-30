'use client';

import { Card } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { useWallet } from '@/hooks/useWallet';
import { useTradeStore } from '@/stores/tradeStore';

export function PortfolioCard() {
  const { wallet, loading } = useWallet();
  const livePrice = useTradeStore((state) => state.livePrice);

  const calculateTotalValue = () => {
    if (!wallet) return 0;
    const btcValue = parseFloat(wallet.btc_balance || '0') * parseFloat(livePrice || '0');
    return parseFloat(wallet.balance || '0') + btcValue;
  };

  if (loading || !wallet) {
    return (
      <Card className="p-6">
        <Skeleton className="h-6 w-32 mb-2" />
        <Skeleton className="h-10 w-48 mb-6" />
        <div className="flex gap-4">
          <div>
            <Skeleton className="h-4 w-16 mb-1" />
            <Skeleton className="h-5 w-24" />
          </div>
          <div>
            <Skeleton className="h-4 w-16 mb-1" />
            <Skeleton className="h-5 w-24" />
          </div>
        </div>
      </Card>
    );
  }

  return (
    <Card className="p-6">
      <h3 className="text-sm font-medium text-[var(--text-secondary)]">Tahmini Bakiye</h3>
      <div className="mt-2 text-4xl font-bold text-[var(--text-primary)]">
        ${calculateTotalValue().toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
      </div>
      
      <div className="mt-6 flex gap-8">
        <div>
          <div className="text-sm text-[var(--text-secondary)]">Kullanılabilir (USD)</div>
          <div className="font-medium text-[var(--text-primary)]">${parseFloat(wallet.balance || '0').toLocaleString()}</div>
        </div>
        <div>
          <div className="text-sm text-[var(--text-secondary)]">Varlık (BTC)</div>
          <div className="font-medium text-[var(--text-primary)]">{parseFloat(wallet.btc_balance || '0')} BTC</div>
        </div>
      </div>
    </Card>
  );
}
