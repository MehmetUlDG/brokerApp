import { useEffect } from 'react';
import { useTradeStore } from '@/stores/tradeStore';
import { BinanceWS } from '@/lib/ws/binance';

export function useLivePrice() {
  const { setLivePrice, addPricePoint } = useTradeStore();

  useEffect(() => {
    const ws = new BinanceWS();
    
    ws.connect((price) => {
      setLivePrice(price);
      addPricePoint({ time: Date.now() / 1000, value: parseFloat(price) });
    });

    return () => {
      ws.disconnect();
    };
  }, []);
}
