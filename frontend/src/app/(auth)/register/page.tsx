import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { RegisterForm } from '@/components/auth/RegisterForm';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Kayıt Ol | TradeOff',
  description: 'TradeOff hesabınızı oluşturun.',
};

export default function RegisterPage() {
  return (
    <Card className="w-full max-w-md p-8 shadow-xl">
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold text-[var(--text-primary)]">Hesap Oluştur</h1>
        <p className="mt-2 text-[var(--text-secondary)]">Kripto dünyasına ilk adımınızı atın</p>
      </div>

      <RegisterForm />

      <div className="mt-6 text-center text-sm text-[var(--text-secondary)]">
        Zaten hesabınız var mı?{' '}
        <Link href="/login" className="font-semibold text-[var(--accent-primary)] hover:underline">
          Giriş Yapın
        </Link>
      </div>
    </Card>
  );
}
