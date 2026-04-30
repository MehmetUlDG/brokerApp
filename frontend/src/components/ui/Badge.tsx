import { HTMLAttributes } from 'react';
import { cn } from '@/lib/utils/cn';

interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'success' | 'danger' | 'warning' | 'info' | 'neutral';
}

export function Badge({ className, variant = 'neutral', children, ...props }: BadgeProps) {
  const variants = {
    success: 'bg-[var(--success)]/10 text-[var(--success)]',
    danger: 'bg-[var(--danger)]/10 text-[var(--danger)]',
    warning: 'bg-[var(--warning)]/10 text-[var(--warning)]',
    info: 'bg-[var(--accent-primary)]/10 text-[var(--accent-primary)]',
    neutral: 'bg-[var(--bg-tertiary)] text-[var(--text-secondary)]',
  };

  return (
    <div
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors',
        variants[variant],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
