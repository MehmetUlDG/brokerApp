import { apiClient } from './client';
import { Wallet, Transaction } from '@/types/domain';

export const walletApi = {
  getWallet: async (): Promise<Wallet> => {
    const response = await apiClient.get<Wallet>('/api/wallet');
    return response.data;
  },
  deposit: async (amount: string, paymentMethodId: string): Promise<{ transaction_id: string; status: string }> => {
    const response = await apiClient.post<{ transaction_id: string; status: string }>('/api/wallet/deposit', { 
      amount, 
      payment_method_id: paymentMethodId 
    });
    return response.data;
  },
  withdraw: async (amount: string, stripeAccountId: string): Promise<{ transaction_id: string; status: string }> => {
    const response = await apiClient.post<{ transaction_id: string; status: string }>('/api/wallet/withdraw', { 
      amount,
      stripe_account_id: stripeAccountId
    });
    return response.data;
  },
  getTransactions: async (): Promise<Transaction[]> => {
    const response = await apiClient.get<{ transactions: Transaction[] }>('/api/transactions');
    return response.data.transactions;
  },
  getBalance: async (): Promise<{ usd_balance: string; btc_balance: string }> => {
    const response = await apiClient.get<{ usd_balance: string; btc_balance: string }>('/api/balance');
    return response.data;
  }
};
