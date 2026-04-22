import { useRef, useCallback } from 'react';
import {
  useMotionValue,
  useSpring,
  useTransform,
} from 'framer-motion';

interface ParallaxTiltOptions {
  maxRotation?: number;  // degrees
  stiffness?: number;
  damping?: number;
}

export function useParallaxTilt(options: ParallaxTiltOptions = {}) {
  const { maxRotation = 12, stiffness = 150, damping = 20 } = options;

  const containerRef = useRef<HTMLDivElement>(null);

  // Raw motion values (normalized -0.5 → 0.5)
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);

  // Glow position (raw 0→1)
  const glowX = useMotionValue(50);
  const glowY = useMotionValue(50);

  // Spring-smoothed values
  const springConfig = { stiffness, damping, mass: 0.8 };
  const smoothX = useSpring(mouseX, springConfig);
  const smoothY = useSpring(mouseY, springConfig);

  // Map to rotation angles
  const rotateY = useTransform(smoothX, [-0.5, 0.5], [-maxRotation, maxRotation]);
  const rotateX = useTransform(smoothY, [-0.5, 0.5], [maxRotation, -maxRotation]);

  const onMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const el = containerRef.current;
    if (!el) return;

    const rect = el.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;

    mouseX.set(x);
    mouseY.set(y);
    glowX.set(((e.clientX - rect.left) / rect.width) * 100);
    glowY.set(((e.clientY - rect.top) / rect.height) * 100);
  }, [mouseX, mouseY, glowX, glowY]);

  const onMouseLeave = useCallback(() => {
    mouseX.set(0);
    mouseY.set(0);
    glowX.set(50);
    glowY.set(50);
  }, [mouseX, mouseY, glowX, glowY]);

  // Depth layers: multiply raw values by a depth factor
  const getLayerTransform = (depth: number) => ({
    x: useTransform(smoothX, [-0.5, 0.5], [-depth * 20, depth * 20]),
    y: useTransform(smoothY, [-0.5, 0.5], [-depth * 20, depth * 20]),
  });

  return {
    containerRef,
    rotateX,
    rotateY,
    glowX,
    glowY,
    onMouseMove,
    onMouseLeave,
    getLayerTransform,
    smoothX,
    smoothY,
  };
}
