import apiRequest from '@/lib/apiRequest';
import { z } from 'zod';
import { type P2POrder, type P2PTrade, p2pOrderSchema, p2pTradeSchema } from './p2p.schema';

export const p2pService = {
  createOrder: async (data: {
    asset_id: number;
    fiat_currency: string;
    rate: string;
    min_amount: string;
    max_amount: string;
    total_amount: string;
    payment_method: string;
  }): Promise<P2POrder> => {
    const response = await apiRequest.post('/p2p/orders', data);
    return p2pOrderSchema.parse(response.data);
  },

  listOrders: async (): Promise<P2POrder[]> => {
    const response = await apiRequest.get('/p2p/orders');
    return z.array(p2pOrderSchema).parse(response.data || []);
  },

  createTrade: async (orderId: string, data: {
    amount: string;
    fiat_amount: string;
  }): Promise<P2PTrade> => {
    const response = await apiRequest.post(`/p2p/orders/${orderId}/trade`, data);
    return p2pTradeSchema.parse(response.data);
  },

  listMyTrades: async (): Promise<P2PTrade[]> => {
    const response = await apiRequest.get('/p2p/trades');
    return z.array(p2pTradeSchema).parse(response.data || []);
  },

  getTrade: async (tradeId: string): Promise<P2PTrade> => {
    const response = await apiRequest.get(`/p2p/trades/${tradeId}`);
    return p2pTradeSchema.parse(response.data);
  },

  markTradePaid: async (tradeId: string): Promise<void> => {
    await apiRequest.post(`/p2p/trades/${tradeId}/pay`);
  },

  releaseTrade: async (tradeId: string): Promise<void> => {
    await apiRequest.post(`/p2p/trades/${tradeId}/release`);
  },

  cancelTrade: async (tradeId: string): Promise<void> => {
    await apiRequest.post(`/p2p/trades/${tradeId}/cancel`);
  }
};
