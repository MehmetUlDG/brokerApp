import React, { forwardRef, useState, useId } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface InputFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  containerClassName?: string;
}

export const InputField = forwardRef<HTMLInputElement, InputFieldProps>(
  ({ label, error, containerClassName = '', className = '', id: propId, ...rest }, ref) => {
    const autoId = useId();
    const fieldId = propId ?? autoId;
    const [isFocused, setIsFocused] = useState(false);
    const [hasValue, setHasValue] = useState(
      () => Boolean(rest.defaultValue || rest.value)
    );

    const isFloated = isFocused || hasValue;

    return (
      <div className={`relative ${containerClassName}`}>
        {/* Floating label */}
        <label
          htmlFor={fieldId}
          style={{
            position: 'absolute',
            left: '18px',
            top: isFloated ? '-10px' : '50%',
            transform: isFloated ? 'translateY(0)' : 'translateY(-50%)',
            fontSize: isFloated ? '0.7rem' : '0.875rem',
            color: isFloated ? 'var(--accent)' : 'var(--text-muted)',
            background: isFloated ? 'var(--bg-primary)' : 'transparent',
            paddingInline: isFloated ? '6px' : '0',
            pointerEvents: 'none',
            transition: 'all 0.18s cubic-bezier(0.16,1,0.3,1)',
            zIndex: 2,
            fontWeight: isFloated ? 600 : 400,
            letterSpacing: isFloated ? '0.06em' : '0',
          }}
        >
          {label}
        </label>

        {/* Shake wrapper on error */}
        <motion.div
          animate={error ? { x: [-8, 8, -6, 6, -4, 4, 0] } : { x: 0 }}
          transition={{ duration: 0.4, ease: 'easeInOut' }}
        >
          <input
            ref={ref}
            id={fieldId}
            {...rest}
            // Ensure placeholder is empty so floating label CSS logic works
            placeholder=" "
            onFocus={(e) => {
              setIsFocused(true);
              rest.onFocus?.(e);
            }}
            onBlur={(e) => {
              setIsFocused(false);
              setHasValue(e.target.value.length > 0);
              rest.onBlur?.(e);
            }}
            onChange={(e) => {
              setHasValue(e.target.value.length > 0);
              rest.onChange?.(e);
            }}
            style={{
              width: '100%',
              borderRadius: '50px',
              border: `1.5px solid ${
                error
                  ? 'rgba(239,68,68,0.8)'
                  : isFocused
                  ? 'var(--accent)'
                  : 'rgba(255,255,255,0.1)'
              }`,
              background: 'rgba(255,255,255,0.04)',
              backdropFilter: 'blur(8px)',
              color: 'var(--text-primary)',
              fontFamily: 'var(--font)',
              padding: '14px 20px',
              fontSize: '0.9rem',
              outline: 'none',
              boxShadow: isFocused
                ? error
                  ? '0 0 0 3px rgba(239,68,68,0.12)'
                  : '0 0 0 3px rgba(141,198,63,0.15), 0 0 20px rgba(141,198,63,0.08)'
                : 'none',
              transition: 'border-color 0.18s ease, box-shadow 0.18s ease',
              ...((rest.style) ?? {}),
            }}
            className={className}
          />
        </motion.div>

        {/* Error message */}
        <AnimatePresence>
          {error && (
            <motion.p
              initial={{ opacity: 0, y: -4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.18 }}
              style={{
                marginTop: '6px',
                marginLeft: '20px',
                fontSize: '0.75rem',
                color: 'rgba(239,68,68,0.9)',
                fontWeight: 500,
              }}
            >
              {error}
            </motion.p>
          )}
        </AnimatePresence>
      </div>
    );
  }
);

InputField.displayName = 'InputField';
