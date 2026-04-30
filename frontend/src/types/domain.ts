export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Wallet {
  id: string;
  user_id: string;
  balance: string;
  btc_balance: string;
  updated_at: string;
}

export type OrderSide = 'BUY' | 'SELL';
export type OrderType = 'MARKET' | 'LIMIT';
export type OrderStatus = 'PENDING' | 'COMPLETED' | 'FAILED' | 'CANCELED';

export interface Order {
  id: string;
  user_id: string;
  symbol: string;
  side: OrderSide;
  type: OrderType;
  quantity: string;
  price: string;
  status: OrderStatus;
  created_at: string;
  updated_at: string;
}

export interface PlaceOrderRequest {
  symbol: string;
  side: OrderSide;
  type: OrderType;
  quantity: string;
  price: string;
}

export interface LivePrice {
  symbol: string;
  price: string;
  timestamp: number;
}

export interface Transaction {
  id: string;
  user_id: string;
  type: 'DEPOSIT' | 'WITHDRAWAL';
  amount: string;
  currency: string;
  status: 'PENDING' | 'COMPLETED' | 'FAILED';
  stripe_ref?: string;
  created_at: string;
}
