import { Metadata } from 'next';
import { PortfolioCard } from '@/components/dashboard/PortfolioCard';
import { PriceCard } from '@/components/dashboard/PriceCard';
import { QuickTradeWidget } from '@/components/dashboard/QuickTradeWidget';
import { RecentOrdersTable } from '@/components/dashboard/RecentOrdersTable';
import { MiniChart } from '@/components/dashboard/MiniChart';

export const metadata: Metadata = {
  title: 'Dashboard | TradeOff',
  description: 'TradeOff hesabınızın genel görünümü.',
};

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[var(--text-primary)]">Genel Bakış</h1>
        <p className="text-[var(--text-secondary)]">Varlıklarınızı ve piyasaları tek bir ekranda takip edin.</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <PortfolioCard />
          <div className="grid gap-6 md:grid-cols-2">
            <PriceCard />
            <MiniChart />
          </div>
          <RecentOrdersTable />
        </div>
        
        <div className="space-y-6">
          <QuickTradeWidget />
        </div>
      </div>
    </div>
  );
}
