import { useWalletStore } from '@/stores/walletStore';
import { useEffect } from 'react';

export function useWallet() {
  const store = useWalletStore();

  useEffect(() => {
    if (!store.wallet && !store.loading) {
      store.fetchWallet();
    }
  }, []);

  return store;
}
