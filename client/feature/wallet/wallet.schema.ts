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
