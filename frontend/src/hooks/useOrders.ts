'use client';

import { useEffect, useCallback } from 'react';
import { useTradeStore } from '@/stores/tradeStore';

const POLL_INTERVAL_MS = 5000; // 5 saniyede bir güncelle

export function useOrders() {
  const store = useTradeStore();

  const refresh = useCallback(() => {
    store.fetchOrders();
  }, [store.fetchOrders]);

  useEffect(() => {
    // İlk yükleme
    store.fetchOrders();

    // 5 saniyede bir polling
    const intervalId = setInterval(() => {
      store.fetchOrders();
    }, POLL_INTERVAL_MS);

    return () => clearInterval(intervalId);
  }, []);

  return { ...store, refresh };
}
