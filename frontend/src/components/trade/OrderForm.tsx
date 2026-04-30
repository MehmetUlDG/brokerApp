'use client';

import { useState } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { toast } from 'sonner';
import { ordersApi } from '@/lib/api/orders';
import { useTradeStore } from '@/stores/tradeStore';
import { useWalletStore } from '@/stores/walletStore';
import { cn } from '@/lib/utils/cn';

export function OrderForm() {
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');
  const [type, setType] = useState<'MARKET' | 'LIMIT'>('MARKET');
  const [quantity, setQuantity] = useState('');
  const [price, setPrice] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const livePrice = useTradeStore((state) => state.livePrice);
  const addOrder = useTradeStore((state) => state.addOrder);
  const fetchWallet = useWalletStore((state) => state.fetchWallet);
  const wallet = useWalletStore((state) => state.wallet);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!quantity || isNaN(Number(quantity)) || Number(quantity) <= 0) {
      toast.error('Lütfen geçerli bir miktar girin.');
      return;
    }
    if (type === 'LIMIT' && (!price || isNaN(Number(price)) || Number(price) <= 0)) {
      toast.error('Lütfen geçerli bir fiyat girin.');
      return;
    }

    setIsSubmitting(true);
    try {
      const order = await ordersApi.placeOrder({
        symbol: 'BTCUSDT',
        side,
        type,
        quantity,
        price: type === 'MARKET' ? (livePrice || '0') : price,
      });
      addOrder(order);
      fetchWallet();
      toast.success(`${side} emri başarıyla iletildi!`);
      setQuantity('');
      if (type === 'LIMIT') setPrice('');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Emir iletilemedi.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const currentPrice = type === 'MARKET' ? livePrice : price;
  const total = parseFloat(quantity || '0') * parseFloat(currentPrice || '0');

  return (
    <Card className="flex flex-col p-6 h-[500px]">
      <div className="flex rounded-md bg-[var(--bg-secondary)] p-1 mb-6">
        <button
          className={cn(
            'flex-1 rounded-sm py-1.5 text-sm font-medium transition-colors',
            side === 'BUY' ? 'bg-[var(--surface)] text-[var(--success)] shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
          )}
          onClick={() => setSide('BUY')}
        >
          Alış
        </button>
        <button
          className={cn(
            'flex-1 rounded-sm py-1.5 text-sm font-medium transition-colors',
            side === 'SELL' ? 'bg-[var(--surface)] text-[var(--danger)] shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-primary)]'
          )}
          onClick={() => setSide('SELL')}
        >
          Satış
        </button>
      </div>

      <div className="flex gap-4 mb-6 text-sm">
        <button
          className={cn('pb-1 font-medium', type === 'MARKET' ? 'border-b-2 border-[var(--accent-primary)] text-[var(--text-primary)]' : 'text-[var(--text-muted)]')}
          onClick={() => setType('MARKET')}
        >
          Piyasa
        </button>
        <button
          className={cn('pb-1 font-medium', type === 'LIMIT' ? 'border-b-2 border-[var(--accent-primary)] text-[var(--text-primary)]' : 'text-[var(--text-muted)]')}
          onClick={() => setType('LIMIT')}
        >
          Limit
        </button>
      </div>

      <form onSubmit={handleSubmit} className="flex-1 flex flex-col justify-between">
        <div className="space-y-4">
          {type === 'LIMIT' && (
            <Input
              label="Fiyat (USDT)"
              type="number"
              step="0.01"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              placeholder="Fiyat girin"
            />
          )}
          <Input
            label="Miktar (BTC)"
            type="number"
            step="0.0001"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder="0.00"
          />
          
          <div className="flex justify-between text-sm py-2 border-t border-[var(--border)] mt-4">
            <span className="text-[var(--text-secondary)]">Tahmini Tutar:</span>
            <span className="font-medium text-[var(--text-primary)]">{total > 0 ? `~${total.toFixed(2)} USDT` : '-'}</span>
          </div>
          
          <div className="flex justify-between text-sm pb-2 border-b border-[var(--border)]">
            <span className="text-[var(--text-secondary)]">Kullanılabilir:</span>
            <span className="font-medium text-[var(--text-primary)]">
              {side === 'BUY' 
                ? `${parseFloat(wallet?.balance || '0').toFixed(2)} USDT` 
                : `${parseFloat(wallet?.btc_balance || '0').toFixed(4)} BTC`}
            </span>
          </div>
        </div>

        <Button
          type="submit"
          className={cn('w-full mt-6', side === 'BUY' ? 'bg-[var(--success)] hover:bg-[var(--success)]/90' : 'bg-[var(--danger)] hover:bg-[var(--danger)]/90')}
          isLoading={isSubmitting}
        >
          {side === 'BUY' ? 'BTC Al' : 'BTC Sat'}
        </Button>
      </form>
    </Card>
  );
}
