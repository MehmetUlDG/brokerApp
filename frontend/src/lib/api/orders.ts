import { apiClient } from './client';
import { PlaceOrderRequest, Order } from '@/types/domain';

export const ordersApi = {
  placeOrder: async (data: PlaceOrderRequest): Promise<Order> => {
    const response = await apiClient.post<Order>('/api/orders', data);
    return response.data;
  },
};
