import { useAuthStore } from '@/stores/authStore';
import { useEffect, useState } from 'react';

export function useAuth() {
  const store = useAuthStore();
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    store.hydrate();
    setIsHydrated(true);
  }, []);

  return { ...store, isHydrated };
}
