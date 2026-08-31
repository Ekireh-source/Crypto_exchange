import { z } from 'zod';
import { assetSchema } from '@/feature/assets/assets.schema';

export const assetBalanceSchema = z.object({
  asset: assetSchema,
  available: z.string().default('0'),
  locked: z.string().default('0'),
  usd_value: z.number().default(0),
});

export type AssetBalance = z.infer<typeof assetBalanceSchema>;

export const portfolioResponseSchema = z.object({
  total_usd_value: z.number().default(0),
  assets: z.array(assetBalanceSchema).default([]),
});

export type PortfolioResponse = z.infer<typeof portfolioResponseSchema>;

export const swapRequestSchema = z.object({
  from_asset_id: z.number(),
  to_asset_id: z.number(),
  amount: z.string(),
});

export type SwapRequest = z.infer<typeof swapRequestSchema>;

export const swapResponseSchema = z.object({
  swap: z.any(), // Adjust type as needed based on backend Swap model
});

export type SwapResponse = z.infer<typeof swapResponseSchema>;

export const transactionSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  asset_id: z.number(),
  type: z.enum(["deposit", "withdrawal", "swap", "p2p_buy", "p2p_sell"]),
  status: z.enum(["pending", "confirmed", "failed"]),
  amount: z.string(),
  fee: z.string(),
  tx_hash: z.string().nullable().optional(),
  from_address: z.string().nullable().optional(),
  to_address: z.string().nullable().optional(),
  note: z.string().nullable().optional(),
  created_at: z.string(),
  confirmed_at: z.string().nullable().optional(),
});

export type Transaction = z.infer<typeof transactionSchema>;

export const transactionResponseSchema = z.object({
  transactions: z.array(transactionSchema),
  total: z.number(),
  page: z.number(),
  limit: z.number(),
});

export type TransactionResponse = z.infer<typeof transactionResponseSchema>;
