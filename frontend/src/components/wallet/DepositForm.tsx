'use client';

import { useState } from 'react';
import { loadStripe } from '@stripe/stripe-js';
import { Elements, CardElement, useStripe, useElements } from '@stripe/react-stripe-js';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { toast } from 'sonner';
import { walletApi } from '@/lib/api/wallet';
import { useWalletStore } from '@/stores/walletStore';
import { STRIPE_PUBLISHABLE_KEY } from '@/lib/constants';
import { useThemeStore } from '@/stores/themeStore';

const stripePromise = loadStripe(STRIPE_PUBLISHABLE_KEY);

function DepositFormContent() {
  const stripe = useStripe();
  const elements = useElements();
  const theme = useThemeStore((state) => state.theme);
  
  const [amount, setAmount] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const fetchWallet = useWalletStore((state) => state.fetchWallet);

  const handleDeposit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!amount || isNaN(Number(amount)) || Number(amount) <= 0) {
      toast.error('Lütfen geçerli bir tutar girin.');
      return;
    }

    if (!stripe || !elements) {
      toast.error('Ödeme sistemi yükleniyor...');
      return;
    }

    setIsSubmitting(true);
    try {
      // In a real flow:
      // 1. Create PaymentIntent on backend via gRPC Proxy.
      // 2. Confirm card payment with stripe.confirmCardPayment using clientSecret.
      // 3. Backend verifies and updates wallet via Webhook.
      
      // Here we simulate successful Stripe flow + Wallet update (Mocking the proxy)
      const cardElement = elements.getElement(CardElement);
      if (cardElement) {
        const { error, paymentMethod } = await stripe.createPaymentMethod({
          type: 'card',
          card: cardElement,
        });

        if (error) {
          throw new Error(error.message);
        }
      }
      
      await walletApi.deposit(amount);
      fetchWallet();
      toast.success('Para yatırma işlemi başarıyla gerçekleştirildi!');
      setAmount('');
      elements.getElement(CardElement)?.clear();
    } catch (error: any) {
      toast.error(error.message || error.response?.data?.error || 'Para yatırma işlemi başarısız oldu.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const cardStyle = {
    style: {
      base: {
        color: theme === 'dark' ? '#F9FAFB' : '#111827',
        fontFamily: '"Inter", sans-serif',
        fontSmoothing: 'antialiased',
        fontSize: '14px',
        '::placeholder': {
          color: theme === 'dark' ? '#9CA3AF' : '#6B7280',
        },
      },
      invalid: {
        color: '#EF4444',
        iconColor: '#EF4444',
      },
    },
  };

  return (
    <form onSubmit={handleDeposit} className="space-y-4">
      <Input
        label="Yatırılacak Tutar (USD)"
        type="number"
        step="0.01"
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
        placeholder="0.00"
      />
      
      <div className="space-y-1">
        <label className="text-sm font-medium text-[var(--text-primary)]">Kart Bilgileri</label>
        <div className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-3 py-3">
          <CardElement options={cardStyle} />
        </div>
      </div>

      <Button type="submit" className="w-full" isLoading={isSubmitting} disabled={!stripe}>
        Para Yatır
      </Button>
    </form>
  );
}

export function DepositForm() {
  return (
    <Elements stripe={stripePromise}>
      <DepositFormContent />
    </Elements>
  );
}
