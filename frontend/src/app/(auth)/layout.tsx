import Link from 'next/link';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col bg-[var(--bg-primary)]">
      <header className="flex h-16 items-center justify-between px-6 border-b border-[var(--border)] bg-[var(--surface)]">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded bg-[var(--accent-primary)] font-bold text-white">
            T
          </div>
          <span className="text-xl font-bold text-[var(--text-primary)]">TradeOff</span>
        </Link>
        <ThemeToggle />
      </header>
      
      <main className="flex flex-1 items-center justify-center p-4">
        {children}
      </main>
    </div>
  );
}
