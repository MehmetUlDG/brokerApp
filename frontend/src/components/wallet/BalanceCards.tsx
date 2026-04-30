'use client';

import { Card } from '@/components/ui/Card';
import { useWallet } from '@/hooks/useWallet';
import { useTradeStore } from '@/stores/tradeStore';
import { Skeleton } from '@/components/ui/Skeleton';
import { Wallet, Bitcoin } from 'lucide-react';

export function BalanceCards() {
  const { wallet, loading } = useWallet();
  const livePrice = useTradeStore((state) => state.livePrice);

  if (loading || !wallet) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="p-6">
          <Skeleton className="h-10 w-10 rounded-full mb-4" />
          <Skeleton className="h-6 w-32 mb-2" />
          <Skeleton className="h-10 w-48" />
        </Card>
        <Card className="p-6">
          <Skeleton className="h-10 w-10 rounded-full mb-4" />
          <Skeleton className="h-6 w-32 mb-2" />
          <Skeleton className="h-10 w-48" />
        </Card>
      </div>
    );
  }

  const usdBalance = parseFloat(wallet.balance || '0');
  const btcBalance = parseFloat(wallet.btc_balance || '0');
  const btcUsdValue = btcBalance * parseFloat(livePrice || '0');

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <Card className="p-6 bg-gradient-to-br from-[var(--surface)] to-[var(--bg-secondary)]">
        <div className="flex items-center gap-4 mb-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[var(--success)]/10 text-[var(--success)]">
            <Wallet className="h-6 w-6" />
          </div>
          <div>
            <h3 className="font-bold text-[var(--text-primary)]">USD Bakiyesi</h3>
            <p className="text-sm text-[var(--text-secondary)]">Kullanılabilir Nakit</p>
          </div>
        </div>
        <div className="mt-4">
          <span className="text-4xl font-bold text-[var(--text-primary)]">${usdBalance.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
        </div>
      </Card>

      <Card className="p-6 bg-gradient-to-br from-[var(--surface)] to-[#F7931A]/5 border-[#F7931A]/20">
        <div className="flex items-center gap-4 mb-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#F7931A]/10 text-[#F7931A]">
            <Bitcoin className="h-6 w-6" />
          </div>
          <div>
            <h3 className="font-bold text-[var(--text-primary)]">BTC Varlığı</h3>
            <p className="text-sm text-[var(--text-secondary)]">Bitcoin</p>
          </div>
        </div>
        <div className="mt-4 flex flex-col">
          <span className="text-4xl font-bold text-[var(--text-primary)]">{btcBalance} <span className="text-2xl text-[var(--text-secondary)]">BTC</span></span>
          <span className="text-sm text-[var(--text-secondary)] mt-1">≈ ${btcUsdValue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
        </div>
      </Card>
    </div>
  );
}
