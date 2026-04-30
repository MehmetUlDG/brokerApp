import { ReactNode } from 'react';
import { Card } from '@/components/ui/Card';

interface FeatureCardProps {
  title: string;
  description: string;
  icon: ReactNode;
}

export function FeatureCard({ title, description, icon }: FeatureCardProps) {
  return (
    <Card className="group overflow-hidden p-6 transition-all hover:shadow-md hover:border-[var(--accent-primary)]/50">
      <div className="mb-4 inline-flex rounded-lg bg-[var(--accent-primary)]/10 p-3 text-[var(--accent-primary)] group-hover:bg-[var(--accent-primary)] group-hover:text-white transition-colors">
        {icon}
      </div>
      <h3 className="mb-2 text-xl font-bold text-[var(--text-primary)]">{title}</h3>
      <p className="text-[var(--text-secondary)] leading-relaxed">
        {description}
      </p>
    </Card>
  );
}
