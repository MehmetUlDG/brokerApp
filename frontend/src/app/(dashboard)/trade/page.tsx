import { Metadata } from 'next';
import { TickerBar } from '@/components/trade/TickerBar';
import { PriceChart } from '@/components/trade/PriceChart';
import { OrderForm } from '@/components/trade/OrderForm';
import { OrderHistory } from '@/components/trade/OrderHistory';

export const metadata: Metadata = {
  title: 'Alım Satım | TradeOff',
  description: 'Kripto para alım satım platformu.',
};

export default function TradePage() {
  return (
    <div className="h-full flex flex-col gap-6">
      <TickerBar />
      
      <div className="grid gap-6 lg:grid-cols-3 flex-1">
        <div className="lg:col-span-2 space-y-6">
          <PriceChart />
          <OrderHistory />
        </div>
        
        <div className="space-y-6">
          <OrderForm />
        </div>
      </div>
    </div>
  );
}
