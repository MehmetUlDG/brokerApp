'use client';

import { useEffect, useRef } from 'react';
import { Card } from '@/components/ui/Card';
import { createChart, IChartApi, ISeriesApi, AreaSeries } from 'lightweight-charts';
import { useTradeStore } from '@/stores/tradeStore';
import { useThemeStore } from '@/stores/themeStore';

export function PriceChart() {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<"Area"> | null>(null);
  
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
        vertLines: { color: theme === 'dark' ? '#2A3451' : '#E5E7EB' },
        horzLines: { color: theme === 'dark' ? '#2A3451' : '#E5E7EB' },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
      },
      rightPriceScale: {
        borderVisible: false,
      },
    };

    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: chartContainerRef.current.clientHeight,
      ...chartOptions
    });

    const series = chart.addSeries(AreaSeries, {
      lineColor: '#3B82F6',
      topColor: 'rgba(59, 130, 246, 0.4)',
      bottomColor: 'rgba(59, 130, 246, 0.0)',
      lineWidth: 2,
    });

    chartRef.current = chart;
    seriesRef.current = series;

    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({ 
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight
        });
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
      const uniqueData = priceHistory.reduce((acc, curr) => {
        if (acc.length === 0 || acc[acc.length - 1].time !== curr.time) {
          acc.push({ time: curr.time as any, value: curr.value });
        }
        return acc;
      }, [] as any[]);

      try {
        seriesRef.current.setData(uniqueData);
      } catch (e) {
        console.error('Lightweight Charts error', e);
      }
    }
  }, [priceHistory]);

  return (
    <Card className="h-[500px] w-full p-1 overflow-hidden">
      <div ref={chartContainerRef} className="h-full w-full" />
    </Card>
  );
}
