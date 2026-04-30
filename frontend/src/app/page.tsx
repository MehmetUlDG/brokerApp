import { Navbar } from "@/components/layout/Navbar";
import { Footer } from "@/components/layout/Footer";
import { HeroSection } from "@/components/landing/HeroSection";
import { FeatureCard } from "@/components/landing/FeatureCard";
import { StatsSection } from "@/components/landing/StatsSection";
import { CTASection } from "@/components/landing/CTASection";
import { Shield, Zap, TrendingUp, RefreshCw } from "lucide-react";

export default function LandingPage() {
  const features = [
    {
      title: "Güvenilir Altyapı",
      description: "Varlıklarınız soğuk cüzdanlarda güvenle saklanır. Gelişmiş şifreleme ile tam koruma.",
      icon: <Shield className="h-6 w-6" />
    },
    {
      title: "Işık Hızında İşlemler",
      description: "Gelişmiş eşleştirme motorumuz sayesinde milisaniyeler içinde işlemleriniz gerçekleşir.",
      icon: <Zap className="h-6 w-6" />
    },
    {
      title: "Gelişmiş Analiz",
      description: "Profesyonel grafikler ve teknik analiz araçları ile piyasayı bir adım önde takip edin.",
      icon: <TrendingUp className="h-6 w-6" />
    },
    {
      title: "7/24 Kesintisiz",
      description: "Binance likiditesi ve sağlam mikroservis mimarimizle her an kesintisiz erişim.",
      icon: <RefreshCw className="h-6 w-6" />
    }
  ];

  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />
      <main className="flex-1">
        <HeroSection />
        
        <section id="features" className="py-24 bg-[var(--bg-primary)]">
          <div className="container mx-auto px-4">
            <div className="text-center mb-16">
              <h2 className="text-3xl font-bold tracking-tight text-[var(--text-primary)] sm:text-4xl">
                Neden TradeOff?
              </h2>
              <p className="mt-4 text-lg text-[var(--text-secondary)]">
                Hem yeni başlayanlar hem de profesyoneller için tasarlandı.
              </p>
            </div>
            <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-4">
              {features.map((f, i) => (
                <FeatureCard key={i} title={f.title} description={f.description} icon={f.icon} />
              ))}
            </div>
          </div>
        </section>

        <StatsSection />
        <CTASection />
      </main>
      <Footer />
    </div>
  );
}
