'use client';

import { useWalletStore } from '@/stores/walletStore';
import { useEffect } from 'react';

const POLL_INTERVAL_MS = 10000; // 10 saniyede bir bakiyeyi güncelle

export function useWallet() {
  const store = useWalletStore();

  useEffect(() => {
    // Her zaman ilk yüklemede fetch et
    store.fetchWallet();

    // Bakiyeyi periyodik olarak güncelle
    const intervalId = setInterval(() => {
      store.fetchWallet();
    }, POLL_INTERVAL_MS);

    return () => clearInterval(intervalId);
  }, []);

  return store;
}
