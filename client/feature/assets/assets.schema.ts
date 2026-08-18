import { z } from 'zod';

export const assetSchema = z.object({
  id: z.number(),
  symbol: z.string(),
  name: z.string(),
  network: z.string(),
  standard: z.string(),
  contract_address: z.string().nullable().optional(),
  decimals: z.number(),
  is_active: z.boolean(),
  logo_url: z.string().optional(),
});

export const assetsResponseSchema = z.array(assetSchema);

export type Asset = z.infer<typeof assetSchema>;
