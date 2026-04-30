import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { AlertCircle } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-[var(--bg-primary)] p-4 text-center">
      <div className="flex h-20 w-20 items-center justify-center rounded-full bg-[var(--danger)]/10 text-[var(--danger)] mb-8">
        <AlertCircle className="h-10 w-10" />
      </div>
      <h1 className="text-8xl font-black text-[var(--text-primary)]">404</h1>
      <h2 className="mt-4 text-2xl font-bold text-[var(--text-primary)]">Sayfa Bulunamadı</h2>
      <p className="mt-2 max-w-md text-[var(--text-secondary)]">
        Aradığınız sayfa silinmiş, adı değiştirilmiş veya geçici olarak kullanılamıyor olabilir.
      </p>
      <Link href="/" className="mt-8">
        <Button size="lg">Ana Sayfaya Dön</Button>
      </Link>
    </div>
  );
}
