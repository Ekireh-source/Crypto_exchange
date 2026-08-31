import { useEffect, useState, useRef } from 'react';

export function usePriceWebsocket() {
  const [prices, setPrices] = useState<Record<string, number>>({});
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let reconnectTimer: NodeJS.Timeout;

    const connect = () => {
      const ws = new WebSocket('ws://localhost:8080/v1/ws/prices');
      
      ws.onmessage = (event) => {
        try {
          const newPrices = JSON.parse(event.data);
          setPrices(newPrices);
        } catch (err) {
          console.error("Failed to parse websocket message", err);
        }
      };

      ws.onclose = () => {
        // Automatically reconnect after 2 seconds
        reconnectTimer = setTimeout(() => {
          connect();
        }, 2000);
      };

      ws.onerror = (err) => {
        console.warn("WebSocket error", err);
      };

      wsRef.current = ws;
    };

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return prices;
}
