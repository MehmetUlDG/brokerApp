import { apiClient } from './client';
import { Wallet, Transaction } from '@/types/domain';

export const walletApi = {
  getWallet: async (): Promise<Wallet> => {
    const response = await apiClient.get<Wallet>('/api/wallet');
    return response.data;
  },
  deposit: async (amount: string): Promise<Wallet> => {
    const response = await apiClient.post<Wallet>('/api/wallet/deposit', { amount });
    return response.data;
  },
  withdraw: async (amount: string): Promise<Wallet> => {
    const response = await apiClient.post<Wallet>('/api/wallet/withdraw', { amount });
    return response.data;
  },
  getTransactions: async (): Promise<Transaction[]> => {
    // Note: Mocking gRPC Proxy /api/transactions
    const response = await apiClient.get<{ transactions: Transaction[] }>('/api/transactions').catch(() => ({ data: { transactions: [] } }));
    return response.data.transactions;
  },
  getBalance: async (): Promise<{ usd_balance: string; btc_balance: string }> => {
    // Note: Mocking gRPC Proxy /api/balance
    const response = await apiClient.get<{ usd_balance: string; btc_balance: string }>('/api/balance').catch(() => ({ data: { usd_balance: '0.00', btc_balance: '0.00' } }));
    return response.data;
  }
};
