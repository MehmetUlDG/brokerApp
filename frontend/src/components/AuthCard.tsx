import React from 'react';
import { motion, type Variants } from 'framer-motion';
import { Link } from 'react-router-dom';

interface AuthCardProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
}

// ─── Tradex Logo ─────────────────────────────────────────────────────────────
const TradexLogo: React.FC = () => (
  <motion.div
    initial={{ opacity: 0, y: -12 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
    style={{ display: 'flex', justifyContent: 'center', marginBottom: '32px' }}
  >
    <Link
      to="/"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        fontWeight: 800,
        fontSize: '1.5rem',
        color: '#8DC63F',
        letterSpacing: '-0.02em',
        textDecoration: 'none',
      }}
    >
      <motion.svg
        width="28"
        height="28"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        whileHover={{ rotate: 12, scale: 1.12 }}
        transition={{ type: 'spring', stiffness: 300, damping: 20 }}
      >
        <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" />
        <path d="M2 17L12 22L22 17" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M2 12L12 17L22 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      </motion.svg>
      TRADEX
    </Link>
  </motion.div>
);

// ─── Variants (explicitly typed to avoid TS inference errors) ─────────────────
const cardVariants: Variants = {
  hidden: { opacity: 0, y: 36, scale: 0.97 },
  visible: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: {
      duration: 0.55,
      ease: [0.16, 1, 0.3, 1],
      staggerChildren: 0.09,
    },
  },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 18 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.45, ease: [0.16, 1, 0.3, 1] },
  },
};

// ─── AuthCard ─────────────────────────────────────────────────────────────────
export const AuthCard: React.FC<AuthCardProps> = ({ children, title, subtitle }) => {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px',
        background:
          'radial-gradient(ellipse 80% 60% at 50% -10%, rgba(141,198,63,0.09) 0%, transparent 70%)',
      }}
    >
      <TradexLogo />

      <motion.div
        variants={cardVariants}
        initial="hidden"
        animate="visible"
        style={{
          width: '100%',
          maxWidth: '440px',
          borderRadius: '24px',
          padding: '40px',
          position: 'relative',
          background: 'rgba(255,255,255,0.04)',
          backdropFilter: 'blur(24px)',
          WebkitBackdropFilter: 'blur(24px)',
          border: '1px solid rgba(255,255,255,0.09)',
          boxShadow: '0 24px 64px rgba(0,0,0,0.5), 0 0 0 1px rgba(141,198,63,0.04)',
        }}
      >
        {/* Top-edge glow line */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: '20%',
            right: '20%',
            height: '1px',
            background:
              'linear-gradient(90deg, transparent, rgba(141,198,63,0.5), transparent)',
            borderRadius: '50%',
          }}
        />

        {/* Title */}
        <motion.h2
          variants={itemVariants}
          style={{
            fontSize: 'clamp(1.8rem, 4vw, 2.25rem)',
            fontWeight: 800,
            color: '#fff',
            letterSpacing: '-0.025em',
            marginBottom: '6px',
            lineHeight: 1.15,
          }}
        >
          {title}
        </motion.h2>

        {/* Subtitle */}
        <motion.p
          variants={itemVariants}
          style={{
            color: '#888',
            fontSize: '0.9rem',
            marginBottom: '32px',
          }}
        >
          {subtitle}
        </motion.p>

        {/* Form content */}
        <motion.div variants={itemVariants}>{children}</motion.div>
      </motion.div>
    </div>
  );
};

// Export for reuse
export { itemVariants };
