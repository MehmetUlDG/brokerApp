import { apiClient } from './client';
import { PlaceOrderRequest, Order } from '@/types/domain';

export const ordersApi = {
  placeOrder: async (data: PlaceOrderRequest): Promise<Order> => {
    const response = await apiClient.post<Order>('/api/orders', data);
    return response.data;
  },
  getOrders: async (): Promise<Order[]> => {
    const response = await apiClient.get<Order[]>('/api/orders');
    return response.data;
  },
};
