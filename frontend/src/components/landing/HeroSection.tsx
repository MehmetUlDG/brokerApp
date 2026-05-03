import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { ArrowRight, BarChart2 } from 'lucide-react';

export function HeroSection() {
  return (
    <section className="relative overflow-hidden bg-[var(--bg-primary)] pt-24 pb-32">
      <div className="absolute inset-0 bg-repeat opacity-[0.03] pointer-events-none" style={{ backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E")` }}></div>
      <div className="absolute top-0 right-0 -translate-y-12 translate-x-1/3">
        <div className="h-[500px] w-[500px] rounded-full bg-[var(--accent-primary)]/20 blur-[120px]"></div>
      </div>
      <div className="absolute bottom-0 left-0 translate-y-1/3 -translate-x-1/3">
        <div className="h-[400px] w-[400px] rounded-full bg-[var(--success)]/20 blur-[100px]"></div>
      </div>

      <div className="container relative mx-auto px-4 text-center">
        <div className="mx-auto flex max-w-fit items-center justify-center space-x-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] px-4 py-1.5 mb-8">
          <span className="flex h-2 w-2 rounded-full bg-[var(--success)]"></span>
          <span className="text-sm font-medium text-[var(--text-secondary)]">Sistem Aktif ve Sorunsuz Çalışıyor</span>
        </div>

        <h1 className="mx-auto max-w-4xl text-5xl font-extrabold tracking-tight text-[var(--text-primary)] sm:text-7xl">
          Geleceğin Kripto Borsasına <span className="text-transparent bg-clip-text bg-gradient-to-r from-[var(--accent-primary)] to-[var(--success)]">Hoş Geldiniz</span>
        </h1>
        
        <p className="mx-auto mt-6 max-w-2xl text-lg text-[var(--text-secondary)] sm:text-xl">
          TradeOff ile kripto varlıklarınızı güvenle yönetin, saniyeler içinde alım satım yapın ve piyasayı gerçek zamanlı takip edin. Profesyonel araçlar, düşük komisyonlar.
        </p>

        <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
          <Link href="/register">
            <Button size="lg" className="w-full sm:w-auto group">
              Hemen Başla
              <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
            </Button>
          </Link>
          <Link href="/trade">
            <Button variant="outline" size="lg" className="w-full sm:w-auto">
              <BarChart2 className="mr-2 h-4 w-4" />
              Piyasaları İncele
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}
