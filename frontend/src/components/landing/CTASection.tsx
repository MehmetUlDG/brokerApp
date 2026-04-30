import Link from 'next/link';
import { Button } from '@/components/ui/Button';

export function CTASection() {
  return (
    <section className="py-24">
      <div className="container mx-auto px-4">
        <div className="relative overflow-hidden rounded-3xl bg-[var(--surface)] border border-[var(--border)] px-6 py-16 text-center shadow-lg sm:px-16">
          <div className="absolute inset-0 bg-gradient-to-br from-[var(--accent-primary)]/10 to-transparent"></div>
          <div className="relative z-10">
            <h2 className="mx-auto max-w-2xl text-3xl font-bold tracking-tight text-[var(--text-primary)] sm:text-4xl">
              Kripto dünyasına ilk adımınızı atın
            </h2>
            <p className="mx-auto mt-4 max-w-xl text-lg text-[var(--text-secondary)]">
              Sadece birkaç dakika içinde hesabınızı oluşturun ve güvenle işlem yapmaya başlayın.
            </p>
            <div className="mt-8 flex justify-center">
              <Link href="/register">
                <Button size="lg" className="rounded-full px-8">
                  Ücretsiz Hesap Oluştur
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
