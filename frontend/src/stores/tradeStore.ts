import { create } from 'zustand';
import { Order } from '@/types/domain';

interface PricePoint {
  time: number;
  value: number;
}

interface TradeState {
  livePrice: string;
  priceHistory: PricePoint[];
  orders: Order[];
  setLivePrice: (price: string) => void;
  addPricePoint: (point: PricePoint) => void;
  addOrder: (order: Order) => void;
  setOrders: (orders: Order[]) => void;
}

export const useTradeStore = create<TradeState>((set) => ({
  livePrice: '0.00',
  priceHistory: [],
  orders: [],
  setLivePrice: (price) => set({ livePrice: price }),
  addPricePoint: (point) => set((state) => ({ 
    priceHistory: [...state.priceHistory.slice(-100), point] 
  })),
  addOrder: (order) => set((state) => ({ orders: [order, ...state.orders] })),
  setOrders: (orders) => set({ orders }),
}));
