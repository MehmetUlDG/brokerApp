import { Metadata } from 'next';
import { BalanceCards } from '@/components/wallet/BalanceCards';
import { DepositForm } from '@/components/wallet/DepositForm';
import { WithdrawForm } from '@/components/wallet/WithdrawForm';
import { TransactionHistory } from '@/components/wallet/TransactionHistory';
import { Card } from '@/components/ui/Card';

export const metadata: Metadata = {
  title: 'Cüzdan | TradeOff',
  description: 'Varlıklarınızı yönetin.',
};

export default function WalletPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[var(--text-primary)]">Cüzdanım</h1>
        <p className="text-[var(--text-secondary)]">Bakiyenizi görüntüleyin, para yatırma ve çekme işlemlerinizi gerçekleştirin.</p>
      </div>

      <BalanceCards />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <TransactionHistory />
        </div>
        
        <div className="space-y-6">
          <Card className="p-6">
            <h3 className="font-bold text-[var(--text-primary)] mb-4">Para Yatır (Stripe)</h3>
            <DepositForm />
          </Card>
          
          <Card className="p-6">
            <h3 className="font-bold text-[var(--text-primary)] mb-4">Para Çek</h3>
            <WithdrawForm />
          </Card>
        </div>
      </div>
    </div>
  );
}
