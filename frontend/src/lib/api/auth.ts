import { apiClient } from './client';
import { AuthResponse } from '@/types/domain';

export const authApi = {
  register: async (data: any): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/api/auth/register', data);
    return response.data;
  },
  login: async (data: any): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/api/auth/login', data);
    return response.data;
  },
};
