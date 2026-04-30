import { useEffect, useState } from 'react';
import { useThemeStore } from '@/stores/themeStore';

export function useTheme() {
  const store = useThemeStore();
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    store.hydrate();
    setIsMounted(true);
  }, []);

  return { ...store, isMounted };
}
