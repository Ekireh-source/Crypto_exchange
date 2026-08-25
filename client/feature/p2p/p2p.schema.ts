import { z } from 'zod';

export const p2pOrderSchema = z.object({
  id: z.string(),
  seller_id: z.string(),
  asset_id: z.number(),
  fiat_currency: z.string(),
  rate: z.string(),
  min_amount: z.string(),
  max_amount: z.string(),
  available_amount: z.string(),
  payment_method: z.string(),
  status: z.string(),
  completion_rate: z.number(),
  total_orders: z.number(),
  created_at: z.string(),
});

export const p2pTradeSchema = z.object({
  id: z.string(),
  order_id: z.string(),
  buyer_id: z.string(),
  seller_id: z.string(),
  asset_id: z.number(),
  amount: z.string(),
  fiat_amount: z.string(),
  rate: z.string(),
  status: z.string(),
  escrow_locked: z.boolean(),
  created_at: z.string(),
  paid_at: z.string().nullable().optional(),
  released_at: z.string().nullable().optional(),
});

export type P2POrder = z.infer<typeof p2pOrderSchema>;
export type P2PTrade = z.infer<typeof p2pTradeSchema>;
