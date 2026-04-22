import axios from 'axios';
import { useAuthStore } from '../store/authStore';

const API_URL = import.meta.env.VITE_API_URL || (import.meta.env as any).REACT_APP_API_URL || '';

// Configure standard API instance
export const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request Interceptor: Attach JWT token if available
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response Interceptor: Handle 401 Unauthorized globally
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// --- API Service Models ---

export interface AuthResponse {
  token: string;
  user: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
  };
}

export interface Wallet {
  id: string;
  user_id: string;
  balance: string; // From backend DECIMAL(18,8) is typically returned as string or number
  created_at?: string;
  updated_at?: string;
}

export interface OrderRequest {
  symbol: string;
  side: 'BUY' | 'SELL';
  type: 'MARKET' | 'LIMIT';
  quantity: string;
  price?: string; // Optional for MARKET
}

export interface OrderResponse {
  id: string;
  user_id: string;
  symbol: string;
  side: string;
  type: string;
  status: string;
  quantity: string;
  price: string | null;
  created_at: string;
}

// --- API Endpoints ---

export const AuthService = {
  register: (data: Record<string, unknown>) => api.post<AuthResponse>('/api/auth/register', data),
  login: (data: Record<string, unknown>) => api.post<AuthResponse>('/api/auth/login', data),
};

export const WalletService = {
  getWallet: () => api.get<Wallet>('/api/wallet'),
  deposit: (amount: string) => api.post<Wallet>('/api/wallet/deposit', { amount }),
  withdraw: (amount: string) => api.post<Wallet>('/api/wallet/withdraw', { amount }),
};

export const OrderService = {
  createOrder: (order: OrderRequest) => api.post<OrderResponse>('/api/orders', order),
};

export interface TransactionFilter {
  type?: 'all' | 'revenue' | 'expense';
}

export interface OrderPayload {
  symbol: string;
  side: 'buy' | 'sell';
  type: 'market' | 'limit';
  price?: number;
  quantity: number;
}

export const apiService = {
  // Wallet endpoints
  getBalance: async (): Promise<{ balance: number }> => {
    const response = await api.get('/api/wallet/balance');
    return response.data;
  },

  deposit: async (amount: number): Promise<{ newBalance: number; transactionId: string }> => {
    const response = await api.post('/api/wallet/deposit', { amount });
    return response.data;
  },

  withdraw: async (amount: number): Promise<{ newBalance: number; transactionId: string }> => {
    const response = await api.post('/api/wallet/withdraw', { amount });
    return response.data;
  },

  getTransactions: async (filter?: 'all' | 'revenue' | 'expense') => {
    const params = filter ? { type: filter } : {};
    const response = await api.get('/api/wallet/transactions', { params });
    return response.data;
  },

  // Orders endpoints
  placeOrder: async (order: OrderPayload) => {
    const backendOrder = {
      symbol: order.symbol,
      side: order.side.toUpperCase(),
      type: order.type.toUpperCase(),
      quantity: order.quantity.toString(),
      price: order.price ? order.price.toString() : "",
    };
    const response = await api.post('/api/orders', backendOrder);
    return response.data;
  },

  // BTC snapshot fallback
  getBTCSnapshot: async () => {
    const response = await api.get('/api/btc/latest');
    return response.data;
  }
};
