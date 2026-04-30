import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { LoginForm } from '@/components/auth/LoginForm';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Giriş Yap | TradeOff',
  description: 'TradeOff hesabınıza giriş yapın.',
};

export default function LoginPage() {
  return (
    <Card className="w-full max-w-md p-8 shadow-xl">
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold text-[var(--text-primary)]">Hoş Geldiniz</h1>
        <p className="mt-2 text-[var(--text-secondary)]">Hesabınıza giriş yapın ve ticarete başlayın</p>
      </div>

      <LoginForm />

      <div className="mt-6 text-center text-sm text-[var(--text-secondary)]">
        Hesabınız yok mu?{' '}
        <Link href="/register" className="font-semibold text-[var(--accent-primary)] hover:underline">
          Hemen Kayıt Olun
        </Link>
      </div>
    </Card>
  );
}
