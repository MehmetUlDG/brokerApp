'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Menu, X } from 'lucide-react';
import { Button } from '@/components/ui/Button';

export function MobileMenu() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <Button variant="ghost" size="sm" onClick={() => setIsOpen(true)} className="px-2">
        <Menu className="h-6 w-6" />
      </Button>

      {isOpen && (
        <div className="fixed inset-0 z-50 flex flex-col bg-[var(--surface)]">
          <div className="flex h-16 items-center justify-between border-b border-[var(--border)] px-4">
            <span className="text-xl font-bold text-[var(--text-primary)]">TradeOff</span>
            <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)} className="px-2">
              <X className="h-6 w-6" />
            </Button>
          </div>
          <div className="flex flex-col p-4 gap-4">
            <Link href="/login" onClick={() => setIsOpen(false)}>
              <Button variant="outline" className="w-full">Giriş Yap</Button>
            </Link>
            <Link href="/register" onClick={() => setIsOpen(false)}>
              <Button className="w-full">Kayıt Ol</Button>
            </Link>
          </div>
        </div>
      )}
    </>
  );
}
