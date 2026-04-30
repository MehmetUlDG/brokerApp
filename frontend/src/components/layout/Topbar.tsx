'use client';

import { Bell, Search } from 'lucide-react';
import { ThemeToggle } from '@/components/ui/ThemeToggle';
import { Input } from '@/components/ui/Input';
import { useAuth } from '@/hooks/useAuth';

export function Topbar() {
  const { user } = useAuth();

  return (
    <header className="flex h-16 items-center justify-between border-b border-[var(--border)] bg-[var(--surface)] px-6">
      <div className="flex flex-1 items-center gap-4">
        <div className="hidden max-w-md flex-1 md:block relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[var(--text-muted)]" />
          <Input className="pl-9 bg-[var(--bg-secondary)] border-transparent" placeholder="Sembol ara (Örn: BTC)" />
        </div>
      </div>

      <div className="flex items-center gap-4">
        <ThemeToggle />
        <button className="relative rounded-full p-2 text-[var(--text-secondary)] hover:bg-[var(--bg-secondary)] hover:text-[var(--text-primary)] transition-colors">
          <Bell className="h-5 w-5" />
          <span className="absolute right-1.5 top-1.5 flex h-2 w-2 rounded-full bg-[var(--danger)]"></span>
        </button>
        <div className="flex items-center gap-3 border-l border-[var(--border)] pl-4">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--accent-primary)] text-sm font-medium text-white">
            {user?.first_name?.[0] || 'U'}
          </div>
          <span className="hidden text-sm font-medium text-[var(--text-primary)] md:block">
            {user?.first_name} {user?.last_name}
          </span>
        </div>
      </div>
    </header>
  );
}
