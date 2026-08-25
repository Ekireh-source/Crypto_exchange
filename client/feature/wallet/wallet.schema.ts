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
