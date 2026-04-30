import { create } from 'zustand';
import { Wallet } from '@/types/domain';
import { walletApi } from '@/lib/api/wallet';

interface WalletState {
  wallet: Wallet | null;
  loading: boolean;
  fetchWallet: () => Promise<void>;
  setWallet: (w: Wallet) => void;
}

export const useWalletStore = create<WalletState>((set) => ({
  wallet: null,
  loading: false,
  fetchWallet: async () => {
    set({ loading: true });
    try {
      const wallet = await walletApi.getWallet();
      set({ wallet, loading: false });
    } catch (error) {
      console.error('Failed to fetch wallet', error);
      set({ loading: false });
    }
  },
  setWallet: (wallet) => set({ wallet }),
}));
