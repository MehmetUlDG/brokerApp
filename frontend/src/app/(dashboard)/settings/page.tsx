import { Metadata } from 'next';
import { Card } from '@/components/ui/Card';

export const metadata: Metadata = {
  title: 'Ayarlar | TradeOff',
  description: 'Hesap ayarlarınızı yönetin.',
};

export default function SettingsPage() {
  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-[var(--text-primary)]">Ayarlar</h1>
        <p className="text-[var(--text-secondary)]">Hesap bilgilerinizi ve tercihlerinizi yönetin.</p>
      </div>

      <Card className="p-6">
        <h3 className="font-bold text-lg mb-4 text-[var(--text-primary)]">Profil Bilgileri</h3>
        <p className="text-[var(--text-secondary)]">
          Profil güncelleme işlemleri henüz aktif değildir. Lütfen daha sonra tekrar deneyin.
        </p>
      </Card>
    </div>
  );
}
