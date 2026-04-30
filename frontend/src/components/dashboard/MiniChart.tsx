'use client';

import { useEffect, useRef } from 'react';
import { Card } from '@/components/ui/Card';
import { createChart, IChartApi, ISeriesApi, LineSeries } from 'lightweight-charts';
import { useTradeStore } from '@/stores/tradeStore';
import { useThemeStore } from '@/stores/themeStore';

export function MiniChart() {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<"Line"> | null>(null);
  
  const priceHistory = useTradeStore((state) => state.priceHistory);
  const theme = useThemeStore((state) => state.theme);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chartOptions = {
      layout: {
        background: { color: 'transparent' },
        textColor: theme === 'dark' ? '#9CA3AF' : '#6B7280',
      },
      grid: {
        vertLines: { visible: false },
        horzLines: { visible: false },
      },
      timeScale: { visible: false },
      rightPriceScale: { visible: false },
      crosshair: {
        horzLine: { visible: false },
        vertLine: { visible: false },
      },
      handleScroll: false,
      handleScale: false,
    };

    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: 120,
      ...chartOptions
    });

    const series = chart.addSeries(LineSeries, {
      color: '#3B82F6',
      lineWidth: 2,
      crosshairMarkerVisible: false,
      priceLineVisible: false,
      lastValueVisible: false,
    });

    chartRef.current = chart;
    seriesRef.current = series;

    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth });
      }
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove();
    };
  }, [theme]);

  useEffect(() => {
    if (seriesRef.current && priceHistory.length > 0) {
      // Lightweight charts requires unique time values ascending.
      // Since our points might be very close, we format them.
      const uniqueData = priceHistory.reduce((acc, curr) => {
        if (acc.length === 0 || acc[acc.length - 1].time !== curr.time) {
          acc.push({ time: curr.time as any, value: curr.value });
        }
        return acc;
      }, [] as any[]);

      try {
        seriesRef.current.setData(uniqueData);
      } catch (e) {
        // Handle chart error if times are not strictly ascending
      }
    }
  }, [priceHistory]);

  return (
    <Card className="p-6 overflow-hidden flex flex-col justify-between">
      <h3 className="font-bold text-[var(--text-primary)] mb-2">Trend (Canlı)</h3>
      <div ref={chartContainerRef} className="w-full h-[120px]" />
    </Card>
  );
}
