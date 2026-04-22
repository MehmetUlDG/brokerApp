import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface Transaction {
  id: string;
  type: 'deposit' | 'withdrawal';
  amount: number;
  timestamp: number;
}

interface WalletState {
  balance: number;
  transactions: Transaction[];
  deposit: (amount: number) => void;
  withdraw: (amount: number) => boolean; // Returns true if successful, false if insufficient
}

export const useWalletStore = create<WalletState>()(
  persist(
    (set, get) => ({
      balance: 0,
      transactions: [],
      
      deposit: (amount: number) => {
        set((state) => ({
          balance: state.balance + amount,
          transactions: [
            {
              id: Date.now().toString() + Math.random().toString(36).substring(7),
              type: 'deposit',
              amount,
              timestamp: Date.now(),
            },
            ...state.transactions,
          ],
        }));
      },
      
      withdraw: (amount: number) => {
        const state = get();
        if (state.balance < amount) {
          return false;
        }
        set((state) => ({
          balance: state.balance - amount,
          transactions: [
            {
              id: Date.now().toString() + Math.random().toString(36).substring(7),
              type: 'withdrawal',
              amount,
              timestamp: Date.now(),
            },
            ...state.transactions,
          ],
        }));
        return true;
      },
    }),
    {
      name: 'tradex-wallet-storage',
    }
  )
);
