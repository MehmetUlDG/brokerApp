import { create } from 'zustand';
import { Order } from '@/types/domain';
import { ordersApi } from '@/lib/api/orders';

interface PricePoint {
  time: number;
  value: number;
}

interface TradeState {
  livePrice: string;
  priceHistory: PricePoint[];
  orders: Order[];
  ordersLoading: boolean;
  setLivePrice: (price: string) => void;
  addPricePoint: (point: PricePoint) => void;
  addOrder: (order: Order) => void;
  setOrders: (orders: Order[]) => void;
  fetchOrders: () => Promise<void>;
}

export const useTradeStore = create<TradeState>((set) => ({
  livePrice: '0.00',
  priceHistory: [],
  orders: [],
  ordersLoading: false,
  setLivePrice: (price) => set({ livePrice: price }),
  addPricePoint: (point) => set((state) => ({
    priceHistory: [...state.priceHistory.slice(-100), point]
  })),
  addOrder: (order) => set((state) => ({ orders: [order, ...state.orders] })),
  setOrders: (orders) => set({ orders }),
  fetchOrders: async () => {
    set({ ordersLoading: true });
    try {
      const orders = await ordersApi.getOrders();
      set({ orders, ordersLoading: false });
    } catch (error) {
      console.error('Failed to fetch orders', error);
      set({ ordersLoading: false });
    }
  },
}));
