'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { toast } from 'sonner';
import { walletApi } from '@/lib/api/wallet';
import { useWalletStore } from '@/stores/walletStore';

export function WithdrawForm() {
  const [amount, setAmount] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const fetchWallet = useWalletStore((state) => state.fetchWallet);
  const wallet = useWalletStore((state) => state.wallet);

  const handleWithdraw = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!amount || isNaN(Number(amount)) || Number(amount) <= 0) {
      toast.error('Lütfen geçerli bir tutar girin.');
      return;
    }

    if (wallet && Number(amount) > parseFloat(wallet.balance)) {
      toast.error('Yetersiz bakiye.');
      return;
    }

    setIsSubmitting(true);
    try {
      await walletApi.withdraw(amount);
      fetchWallet();
      toast.success('Para çekme işlemi başarıyla gerçekleştirildi!');
      setAmount('');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Para çekme işlemi başarısız oldu.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleWithdraw} className="space-y-4">
      <Input
        label="Çekilecek Tutar (USD)"
        type="number"
        step="0.01"
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
        placeholder="0.00"
      />
      <div className="flex justify-between text-sm py-2">
        <span className="text-[var(--text-secondary)]">Mevcut Bakiye:</span>
        <span className="font-medium text-[var(--text-primary)]">
          {parseFloat(wallet?.balance || '0').toLocaleString()} USD
        </span>
      </div>
      <Button type="submit" className="w-full" isLoading={isSubmitting}>
        Para Çek
      </Button>
    </form>
  );
}
