import { create } from 'zustand';
import { User, AuthResponse } from '@/types/domain';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (res: AuthResponse) => void;
  logout: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  login: (res) => {
    localStorage.setItem('token', res.token);
    localStorage.setItem('user', JSON.stringify(res.user));
    // Set cookie for middleware
    document.cookie = `auth_token=${res.token}; path=/; max-age=86400; SameSite=Lax`;
    set({ user: res.user, token: res.token, isAuthenticated: true });
  },
  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    // Remove cookie
    document.cookie = 'auth_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    set({ user: null, token: null, isAuthenticated: false });
  },
  hydrate: () => {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        document.cookie = `auth_token=${token}; path=/; max-age=86400; SameSite=Lax`;
        set({ user, token, isAuthenticated: true });
      } catch (e) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      }
    }
  },
}));
