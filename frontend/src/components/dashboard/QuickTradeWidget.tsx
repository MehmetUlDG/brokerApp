'use client';

import { useState } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { toast } from 'sonner';
import { ordersApi } from '@/lib/api/orders';
import { useTradeStore } from '@/stores/tradeStore';
import { useWalletStore } from '@/stores/walletStore';

export function QuickTradeWidget() {
  const [quantity, setQuantity] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const livePrice = useTradeStore((state) => state.livePrice);
  const addOrder = useTradeStore((state) => state.addOrder);
  const fetchWallet = useWalletStore((state) => state.fetchWallet);

  const handleTrade = async (side: 'BUY' | 'SELL') => {
    if (!quantity || isNaN(Number(quantity)) || Number(quantity) <= 0) {
      toast.error('Lütfen geçerli bir miktar girin.');
      return;
    }

    setIsSubmitting(true);
    try {
      const order = await ordersApi.placeOrder({
        symbol: 'BTCUSDT',
        side,
        type: 'MARKET',
        quantity,
        price: livePrice || '0', // MARKET emri ama backend price isteyebilir
      });
      addOrder(order);
      fetchWallet(); // Bakiye güncellemesi
      toast.success(`${side === 'BUY' ? 'Alım' : 'Satım'} emri başarıyla iletildi!`);
      setQuantity('');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'İşlem başarısız.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="p-6">
      <h3 className="mb-4 font-bold text-[var(--text-primary)]">Hızlı İşlem (Market)</h3>
      <div className="space-y-4">
        <Input
          placeholder="Miktar (BTC)"
          type="number"
          step="0.001"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
        />
        <div className="flex gap-4">
          <Button
            className="flex-1 bg-[var(--success)] hover:bg-[var(--success)]/90"
            onClick={() => handleTrade('BUY')}
            isLoading={isSubmitting}
          >
            BUY BTC
          </Button>
          <Button
            className="flex-1 bg-[var(--danger)] hover:bg-[var(--danger)]/90"
            onClick={() => handleTrade('SELL')}
            isLoading={isSubmitting}
          >
            SELL BTC
          </Button>
        </div>
      </div>
    </Card>
  );
}
