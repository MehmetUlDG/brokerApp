import { useState, useEffect, useRef, useCallback } from 'react';
import { apiService } from '../services/api';
import { useTradingStore } from '../store/tradingStore';

export interface BTCData {
  symbol: string;
  price: number;
  change24h: number;
  high24h: number;
  low24h: number;
  volume: number;
  timestamp: string;
}

export type ConnectionStatus = 'connected' | 'reconnecting' | 'disconnected';

export const useWebSocket = (url: string) => {
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const wsRef = useRef<WebSocket | null>(null);
  const retryCountRef = useRef(0);
  const maxRetries = 5;
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    try {
      // Connect to WebSocket
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus('connected');
        retryCountRef.current = 0; // Reset retries on successful connection
        // Stop polling if we are connected
        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current);
          pollIntervalRef.current = null;
        }
      };

      ws.onmessage = (event) => {
        try {
          const parsedData: BTCData = JSON.parse(event.data);
          useTradingStore.getState().setBtcData({
            price: parsedData.price,
            change24h: parsedData.change24h,
            high24h: parsedData.high24h,
            low24h: parsedData.low24h,
            volume: parsedData.volume,
            isLive: true,
          });
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = () => {
        if (retryCountRef.current < maxRetries) {
          setStatus('reconnecting');
          const timeout = Math.pow(2, retryCountRef.current) * 1000;
          setTimeout(() => {
            retryCountRef.current += 1;
            connect();
          }, timeout);
        } else {
          setStatus('disconnected');
          startPollingFallback();
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket Error:', error);
        ws.close(); // Triggers onclose
      };
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      setStatus('disconnected');
      startPollingFallback();
    }
  }, [url]);

  const startPollingFallback = useCallback(() => {
    if (!pollIntervalRef.current) {
      const poll = async () => {
        try {
          const snapshot = await apiService.getBTCSnapshot();
          // Assume snapshot matches BTCData shape
          if (snapshot) {
            useTradingStore.getState().setBtcData({
              price: snapshot.price,
              change24h: snapshot.change24h,
              high24h: snapshot.high24h,
              low24h: snapshot.low24h,
              volume: snapshot.volume,
              isLive: false,
            });
          }
        } catch (error) {
          console.error('REST fallback failed:', error);
        }
      };

      poll(); // Initial poll
      pollIntervalRef.current = setInterval(poll, 5000);
    }
  }, []);

  useEffect(() => {
    if (url) {
      connect();
    } else {
      // If no URL provided, just use REST fallback directly
      setStatus('disconnected');
      startPollingFallback();
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
      }
    };
  }, [url, connect, startPollingFallback]);

  return { status };
};
