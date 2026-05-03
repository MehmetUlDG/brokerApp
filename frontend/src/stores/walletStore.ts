import { create } from 'zustand';
import { Wallet, Transaction } from '@/types/domain';
import { walletApi } from '@/lib/api/wallet';

interface WalletState {
  wallet: Wallet | null;
  transactions: Transaction[];
  loading: boolean;
  transactionsLoading: boolean;
  fetchWallet: () => Promise<void>;
  fetchTransactions: () => Promise<void>;
  setWallet: (w: Wallet) => void;
}

export const useWalletStore = create<WalletState>((set) => ({
  wallet: null,
  transactions: [],
  loading: false,
  transactionsLoading: false,

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

  fetchTransactions: async () => {
    set({ transactionsLoading: true });
    try {
      const transactions = await walletApi.getTransactions();
      set({ transactions, transactionsLoading: false });
    } catch (error) {
      console.error('Failed to fetch transactions', error);
      set({ transactionsLoading: false });
    }
  },

  setWallet: (wallet) => set({ wallet }),
}));
