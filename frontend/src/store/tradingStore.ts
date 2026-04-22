import { create } from 'zustand';
import { apiService, type OrderPayload } from '../services/api';
import toast from 'react-hot-toast';

interface Transaction {
  id: string;
  type: 'deposit' | 'withdrawal' | 'revenue' | 'expense';
  amount: number;
  timestamp: number;
}

interface Order {
  id: string;
  symbol: string;
  side: string;
  price: number;
  quantity: number;
  status: string;
}

interface TradingState {
  // Global App State
  networkError: boolean;
  setNetworkError: (status: boolean) => void;

  // walletSlice
  wallet: {
    balance: number;
    transactions: Transaction[];
    isLoading: boolean;
  };
  fetchBalance: () => Promise<void>;
  fetchTransactions: () => Promise<void>;
  deposit: (amount: number) => Promise<void>;
  withdraw: (amount: number) => Promise<boolean>;

  // btcSlice
  btc: {
    price: number;
    change24h: number;
    high24h: number;
    low24h: number;
    volume: number;
    isLive: boolean;
    lastUpdate: number;
  };
  setBtcData: (data: Partial<TradingState['btc']>) => void;

  // orderSlice
  order: {
    pendingOrder: boolean;
    orderHistory: Order[];
  };
  placeOrder: (payload: OrderPayload) => Promise<boolean>;
  fetchHistory: () => Promise<void>;
}

const handleApiError = (error: any, set: any) => {
  if (!navigator.onLine || error.message === 'Network Error') {
    set({ networkError: true });
    return;
  }
  set({ networkError: false });
  const msg = error?.response?.data?.message || error?.message || 'An error occurred';
  toast.error(msg);
};

export const useTradingStore = create<TradingState>((set, get) => ({
  networkError: !navigator.onLine,
  setNetworkError: (status) => set({ networkError: status }),

  wallet: {
    balance: 0,
    transactions: [],
    isLoading: false,
  },

  fetchBalance: async () => {
    try {
      set((state) => ({ wallet: { ...state.wallet, isLoading: true } }));
      const data = await apiService.getBalance();
      set((state) => ({
        wallet: { ...state.wallet, balance: data.balance, isLoading: false },
        networkError: false,
      }));
    } catch (error) {
      set((state) => ({ wallet: { ...state.wallet, isLoading: false } }));
      handleApiError(error, set);
    }
  },

  fetchTransactions: async () => {
    try {
      const data = await apiService.getTransactions();
      set((state) => ({
        wallet: { ...state.wallet, transactions: data || [] },
      }));
    } catch (error) {
      handleApiError(error, set);
    }
  },

  deposit: async (amount: number) => {
    try {
      set((state) => ({ wallet: { ...state.wallet, isLoading: true } }));
      const res = await apiService.deposit(amount);
      set((state) => ({
        wallet: {
          ...state.wallet,
          balance: res.newBalance,
          transactions: [
            {
              id: res.transactionId || Date.now().toString(),
              type: 'deposit',
              amount,
              timestamp: Date.now(),
            },
            ...state.wallet.transactions,
          ],
          isLoading: false,
        },
      }));
      toast.success('Deposit successful');
    } catch (error) {
      set((state) => ({ wallet: { ...state.wallet, isLoading: false } }));
      handleApiError(error, set);
    }
  },

  withdraw: async (amount: number) => {
    try {
      set((state) => ({ wallet: { ...state.wallet, isLoading: true } }));
      const res = await apiService.withdraw(amount);
      set((state) => ({
        wallet: {
          ...state.wallet,
          balance: res.newBalance,
          transactions: [
            {
              id: res.transactionId || Date.now().toString(),
              type: 'withdrawal',
              amount,
              timestamp: Date.now(),
            },
            ...state.wallet.transactions,
          ],
          isLoading: false,
        },
      }));
      toast.success('Withdrawal successful');
      return true;
    } catch (error) {
      set((state) => ({ wallet: { ...state.wallet, isLoading: false } }));
      handleApiError(error, set);
      return false;
    }
  },

  btc: {
    price: 64230.50,
    change24h: 2.3,
    high24h: 65000,
    low24h: 63000,
    volume: 18432.5,
    isLive: false,
    lastUpdate: Date.now(),
  },

  setBtcData: (data) => set((state) => ({ btc: { ...state.btc, ...data, lastUpdate: Date.now() } })),

  order: {
    pendingOrder: false,
    orderHistory: [],
  },

  placeOrder: async (payload: OrderPayload) => {
    try {
      set((state) => ({ order: { ...state.order, pendingOrder: true } }));
      const res = await apiService.placeOrder(payload);
      toast.success(`Order placed! ID: ${res.id || 'N/A'}`);
      set((state) => ({
        order: { ...state.order, pendingOrder: false },
      }));
      get().fetchBalance(); // Refresh balance after order
      return true;
    } catch (error) {
      set((state) => ({ order: { ...state.order, pendingOrder: false } }));
      handleApiError(error, set);
      return false;
    }
  },

  fetchHistory: async () => {
    // If there's an endpoint for order history, we'd call it here
  },
}));

// Listen to online/offline events to manage networkError state automatically
if (typeof window !== 'undefined') {
  window.addEventListener('online', () => useTradingStore.getState().setNetworkError(false));
  window.addEventListener('offline', () => useTradingStore.getState().setNetworkError(true));
}
