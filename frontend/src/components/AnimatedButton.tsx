import React from 'react';
import { motion, type HTMLMotionProps } from 'framer-motion';

interface AnimatedButtonProps extends Omit<HTMLMotionProps<'button'>, 'children'> {
  children: React.ReactNode;
  isLoading?: boolean;
  variant?: 'primary' | 'secondary';
  fullWidth?: boolean;
}

// Loading spinner SVG
const Spinner: React.FC = () => (
  <motion.svg
    width="18"
    height="18"
    viewBox="0 0 24 24"
    fill="none"
    animate={{ rotate: 360 }}
    transition={{ duration: 0.8, repeat: Infinity, ease: 'linear' }}
    style={{ display: 'inline-block' }}
  >
    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" strokeOpacity="0.25" />
    <path
      d="M12 2 A10 10 0 0 1 22 12"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
    />
  </motion.svg>
);

export const AnimatedButton: React.FC<AnimatedButtonProps> = ({
  children,
  isLoading = false,
  variant = 'primary',
  fullWidth = false,
  disabled,
  style,
  ...rest
}) => {
  const isPrimary = variant === 'primary';

  return (
    <motion.button
      whileHover={!disabled && !isLoading ? {
        scale: 1.025,
        boxShadow: isPrimary
          ? '0 8px 32px rgba(141,198,63,0.45), 0 2px 8px rgba(0,0,0,0.3)'
          : '0 8px 32px rgba(0,0,0,0.4)',
      } : {}}
      whileTap={!disabled && !isLoading ? { scale: 0.96 } : {}}
      transition={{ type: 'spring', stiffness: 400, damping: 25 }}
      disabled={disabled || isLoading}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '8px',
        width: fullWidth ? '100%' : 'auto',
        borderRadius: '50px',
        padding: '14px 28px',
        fontFamily: 'var(--font)',
        fontSize: '0.875rem',
        fontWeight: 700,
        letterSpacing: '0.08em',
        textTransform: 'uppercase',
        cursor: disabled || isLoading ? 'not-allowed' : 'pointer',
        opacity: disabled || isLoading ? 0.65 : 1,
        border: isPrimary ? 'none' : '1.5px solid rgba(141,198,63,0.4)',
        background: isPrimary
          ? 'linear-gradient(135deg, #8DC63F 0%, #a8d84a 100%)'
          : 'rgba(141,198,63,0.06)',
        color: isPrimary ? '#000' : '#8DC63F',
        boxShadow: isPrimary
          ? '0 4px 16px rgba(141,198,63,0.25)'
          : 'none',
        position: 'relative',
        overflow: 'hidden',
        transition: 'background 0.2s ease',
        ...style,
      }}
      {...rest}
    >
      {/* Shimmer overlay on hover */}
      <motion.span
        style={{
          position: 'absolute',
          inset: 0,
          background: 'linear-gradient(105deg, transparent 40%, rgba(255,255,255,0.15) 50%, transparent 60%)',
          backgroundSize: '200% 100%',
          backgroundPosition: '200% 0',
        }}
        whileHover={{ backgroundPosition: '-200% 0' }}
        transition={{ duration: 0.5 }}
      />

      {isLoading ? (
        <>
          <Spinner />
          <span>Loading...</span>
        </>
      ) : (
        children
      )}
    </motion.button>
  );
};
