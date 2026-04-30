import { useTradeStore } from '@/stores/tradeStore';

export function useOrders() {
  const store = useTradeStore();
  return store;
}
