import React, { useState } from 'react';
import { motion, useMotionValue, useTransform } from 'framer-motion';
import toast from 'react-hot-toast';
import { ArrowUpRight, ArrowDownRight, ArrowRight } from 'lucide-react';
import { useWalletStore } from '../store/walletStore';
import { api } from '../services/api';
import { loadStripe } from '@stripe/stripe-js';
import { Elements, PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js';

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLIC_KEY || 'pk_test_YOUR_STRIPE_PUBLIC_KEY');

const CheckoutForm = ({ amount, onSuccess, onCancel }: { amount: number, onSuccess: () => void, onCancel: () => void }) => {
  const stripe = useStripe();
  const elements = useElements();
  const [isProcessing, setIsProcessing] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    setIsProcessing(true);
    const { error, paymentIntent } = await stripe.confirmPayment({
      elements,
      redirect: 'if_required',
    });

    if (error) {
      toast.error(error.message || 'Payment failed');
    } else if (paymentIntent && paymentIntent.status === 'succeeded') {
      onSuccess();
    }
    setIsProcessing(false);
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4 mt-4">
      <PaymentElement />
      <div className="flex gap-2 mt-2">
        <button
          type="button"
          onClick={onCancel}
          className="flex-1 py-3 px-4 bg-muted text-muted-foreground font-bold rounded-lg hover:bg-muted/80 transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={!stripe || isProcessing}
          className="flex-1 py-3 px-4 bg-primary text-primary-foreground font-bold rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50"
        >
          {isProcessing ? 'Processing...' : `Pay $${amount}`}
        </button>
      </div>
    </form>
  );
};

export const WalletSection: React.FC = () => {
  const { balance, transactions, deposit, withdraw } = useWalletStore();
  const [depositAmount, setDepositAmount] = useState<string>('');
  const [withdrawAmount, setWithdrawAmount] = useState<string>('');
  const [clientSecret, setClientSecret] = useState<string | null>(null);
  const [isInitiating, setIsInitiating] = useState(false);
  const [activeTab, setActiveTab] = useState<'all' | 'revenue' | 'expenses'>('all');

  // Format currency
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
    }).format(value);
  };

  // Date formatting
  const formatDate = (timestamp: number) => {
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(timestamp));
  };

  const handleDepositInitiate = async () => {
    const amount = parseFloat(depositAmount);
    if (isNaN(amount) || amount < 0.01) {
      toast.error('Please enter a valid amount (min $0.01)');
      return;
    }
    setIsInitiating(true);
    try {
      const res = await api.post('/api/wallet/deposit-intent', { amount });
      if (res.data && res.data.client_secret) {
        setClientSecret(res.data.client_secret);
      } else {
        toast.error('Failed to get client secret');
      }
    } catch (error) {
      toast.error('Failed to initiate deposit');
    } finally {
      setIsInitiating(false);
    }
  };

  const handleDepositSuccess = () => {
    toast.success('Payment successful!');
    deposit(parseFloat(depositAmount));
    setClientSecret(null);
    setDepositAmount('');
  };

  const handleWithdraw = async () => {
    const amount = parseFloat(withdrawAmount);
    if (isNaN(amount) || amount < 0.01) {
      toast.error('Please enter a valid amount (min $0.01)');
      return;
    }
    const success = await withdraw(amount);
    if (success) {
      setWithdrawAmount('');
    }
  };

  const filteredTransactions = transactions.filter((t) => {
    if (activeTab === 'all') return true;
    if (activeTab === 'revenue') return t.type === 'deposit';
    if (activeTab === 'expenses') return t.type === 'withdrawal';
    return true;
  });

  // Card 3D effect
  const x = useMotionValue(0);
  const y = useMotionValue(0);

  const rotateX = useTransform(y, [-100, 100], [15, -15]);
  const rotateY = useTransform(x, [-100, 100], [-15, 15]);

  const handleMouseMove = (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;
    const mouseX = event.clientX - rect.left;
    const mouseY = event.clientY - rect.top;
    const xPct = mouseX / width - 0.5;
    const yPct = mouseY / height - 0.5;
    x.set(xPct * 200);
    y.set(yPct * 200);
  };

  const handleMouseLeave = () => {
    x.set(0);
    y.set(0);
  };

  return (
    <div className="w-full max-w-6xl mx-auto py-12 px-4">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 items-start">

        {/* Left Side: 3D Wallet Card */}
        <div
          className="relative w-full h-[400px] flex items-center justify-center rounded-3xl overflow-hidden"
          style={{ perspective: 1000 }}
        >
          {/* Card Container */}
          <motion.div
            style={{ rotateX, rotateY, transformStyle: "preserve-3d" }}
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="relative w-full max-w-[420px] aspect-[1.6/1] rounded-2xl cursor-pointer"
            initial={{ rotateY: -15, rotateX: 5 }}
            animate={{ rotateY: x.get() ? rotateY.get() : -15, rotateX: y.get() ? rotateX.get() : 5 }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          >
            {/* The actual visual card based on theme */}
            <div className="absolute inset-0 w-full h-full rounded-2xl p-8 flex flex-col justify-between
              /* Light theme: Solid lime-green */
              bg-[#84CC16] text-black shadow-[10px_20px_40px_rgba(0,0,0,0.1)] border border-[#a3e635]
              /* Dark theme: Glassmorphism with neon green glow */
              dark:bg-black/40 dark:backdrop-blur-md dark:border dark:border-[#84CC16]/60 dark:shadow-[0_0_40px_rgba(132,204,22,0.25)] dark:text-white
            ">
              <div className="absolute top-0 left-0 w-full h-full overflow-hidden rounded-2xl pointer-events-none">
                {/* Decorative circles to mimic card chip / logo */}
                <div className="absolute top-6 right-6 w-12 h-12 rounded-full border border-black/20 dark:border-white/20 flex items-center justify-center">
                  <div className="w-8 h-8 rounded-full bg-black/10 dark:bg-white/10" />
                </div>
                <div className="absolute top-6 right-14 w-12 h-12 rounded-full border border-black/20 dark:border-white/20" />
              </div>

              <div className="relative z-10" style={{ transform: "translateZ(40px)" }}>
                <p className="text-lg font-medium opacity-80 mb-1">Available Balance</p>
                <h1 className="text-5xl font-bold tracking-tight">
                  {formatCurrency(balance)}
                </h1>
              </div>

              <div className="relative z-10 flex justify-between items-end" style={{ transform: "translateZ(30px)" }}>
                <div>
                  <p className="text-sm opacity-70 mb-1">Account ID</p>
                  <p className="text-lg font-mono tracking-widest opacity-90">
                    **** **** 0000
                  </p>
                </div>
              </div>
            </div>

            {/* Simulated thickness / stack for light mode like in image */}
            <div className="absolute inset-0 w-full h-full rounded-2xl bg-[#65A30D] -z-10 translate-y-4 translate-x-2 dark:hidden"></div>
          </motion.div>
        </div>

        {/* Right Side: Panels */}
        <div className="flex flex-col gap-6">

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Deposit Panel */}
            <div className="card p-6 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-xl font-bold">Deposit Funds</h3>
                  <div className="w-8 h-8 rounded-full bg-green-500/10 text-green-500 flex items-center justify-center">
                    <ArrowUpRight className="w-5 h-5" />
                  </div>
                </div>
                <p className="text-sm text-muted-foreground mb-6">Add money to your wallet</p>

                {!clientSecret ? (
                  <>
                    <div className="mb-4">
                      <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 block">
                        Amount (USD)
                      </label>
                      <input
                        type="number"
                        min="0.01"
                        step="0.01"
                        placeholder="$1.00"
                        value={depositAmount}
                        onChange={(e) => setDepositAmount(e.target.value)}
                        className="w-full bg-background border border-border rounded-lg p-3 outline-none focus:border-primary transition-colors text-lg"
                      />
                    </div>
                    <button
                      onClick={handleDepositInitiate}
                      disabled={isInitiating}
                      className="w-full py-3 px-4 bg-primary text-primary-foreground font-bold rounded-lg hover:bg-primary/90 transition-colors flex items-center justify-center gap-2 disabled:opacity-50"
                    >
                      {isInitiating ? 'PLEASE WAIT...' : 'CONFIRM DEPOSIT'} {!isInitiating && <ArrowRight className="w-4 h-4" />}
                    </button>
                  </>
                ) : (
                  <Elements stripe={stripePromise} options={{ clientSecret, appearance: { theme: 'night' } }}>
                    <CheckoutForm 
                      amount={parseFloat(depositAmount)} 
                      onSuccess={handleDepositSuccess} 
                      onCancel={() => setClientSecret(null)} 
                    />
                  </Elements>
                )}
              </div>
            </div>

            {/* Withdraw Panel */}
            <div className="card p-6 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-xl font-bold">Withdraw Funds</h3>
                  <div className="w-8 h-8 rounded-full bg-orange-500/10 text-orange-500 flex items-center justify-center">
                    <ArrowDownRight className="w-5 h-5" />
                  </div>
                </div>
                <p className="text-sm text-muted-foreground mb-6">Transfer money to your bank</p>

                <div className="mb-4">
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 block">
                    Amount (USD)
                  </label>
                  <input
                    type="number"
                    min="0.01"
                    step="0.01"
                    placeholder="$1.00"
                    value={withdrawAmount}
                    onChange={(e) => setWithdrawAmount(e.target.value)}
                    className="w-full bg-background border border-border rounded-lg p-3 outline-none focus:border-primary transition-colors text-lg"
                  />
                </div>
              </div>
              <button
                onClick={handleWithdraw}
                className="w-full py-3 px-4 bg-primary text-primary-foreground font-bold rounded-lg hover:bg-primary/90 transition-colors"
              >
                WITHDRAW FUNDS
              </button>
            </div>
          </div>

          {/* Recent Transactions */}
          <div className="card p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-xl font-bold">Recent Transactions</h3>
              <button className="text-sm text-primary hover:underline font-medium">View All</button>
            </div>

            {/* Tabs */}
            <div className="flex gap-6 border-b border-border mb-4">
              {(['all', 'revenue', 'expenses'] as const).map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className={`pb-2 text-sm font-medium capitalize transition-colors relative ${activeTab === tab ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
                    }`}
                >
                  {tab}
                  {activeTab === tab && (
                    <motion.div
                      layoutId="tab-indicator"
                      className="absolute bottom-0 left-0 w-full h-0.5 bg-primary rounded-t-full"
                    />
                  )}
                </button>
              ))}
            </div>

            {/* List */}
            <div className="flex flex-col gap-4 max-h-[300px] overflow-y-auto pr-2">
              {filteredTransactions.length === 0 ? (
                <div className="py-8 text-center text-muted-foreground">
                  No transactions yet
                </div>
              ) : (
                filteredTransactions.map((tx) => (
                  <div key={tx.id} className="flex items-center justify-between p-3 rounded-lg hover:bg-muted/50 transition-colors">
                    <div className="flex items-center gap-4">
                      <div className={`w-10 h-10 rounded-full flex items-center justify-center ${tx.type === 'deposit' ? 'bg-primary/20 text-primary' : 'bg-destructive/20 text-destructive'
                        }`}>
                        {tx.type === 'deposit' ? <ArrowUpRight className="w-5 h-5" /> : <ArrowDownRight className="w-5 h-5" />}
                      </div>
                      <div>
                        <p className="font-semibold">{tx.type === 'deposit' ? 'Account Deposit' : 'Account Withdrawal'}</p>
                        <p className="text-xs text-muted-foreground">{formatDate(tx.timestamp)}</p>
                      </div>
                    </div>
                    <div className={`font-bold ${tx.type === 'deposit' ? 'text-primary' : 'text-destructive'}`}>
                      {tx.type === 'deposit' ? '+' : '-'} {formatCurrency(tx.amount)}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
};
