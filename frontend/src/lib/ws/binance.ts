import { WS_URL } from '../constants';

export class BinanceWS {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private isConnecting = false;

  connect(onPrice: (price: string) => void) {
    if (this.isConnecting || this.ws?.readyState === WebSocket.OPEN) return;
    
    this.isConnecting = true;
    this.ws = new WebSocket(WS_URL);
    
    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.p) {
          onPrice(data.p);
        }
      } catch (e) {
        console.error('Failed to parse WS message', e);
      }
    };

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.isConnecting = false;
    };

    this.ws.onclose = () => {
      this.isConnecting = false;
      this.reconnect(onPrice);
    };

    this.ws.onerror = () => {
      this.isConnecting = false;
      this.ws?.close();
    };
  }

  private reconnect(onPrice: (price: string) => void) {
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    setTimeout(() => { 
      this.reconnectAttempts++; 
      this.connect(onPrice); 
    }, delay);
  }

  disconnect() { 
    if (this.ws) {
      this.ws.onclose = null; // Prevent reconnect logic
      this.ws.close(); 
      this.ws = null;
    }
  }
}
