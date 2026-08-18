import { z } from 'zod';

export const depositAddressResponseSchema = z.object({
  asset_id: z.number(),
  symbol: z.string(),
  network: z.string(),
  standard: z.string(),
  address: z.string(),
  qr_data: z.string(),
  network_warning: z.string(),
});

export type DepositAddressResponse = z.infer<typeof depositAddressResponseSchema>;

export const sendRequestSchema = z.object({
  asset_id: z.number(),
  to_address: z.string().min(1, 'Recipient address is required'),
  amount: z.string().min(1, 'Amount is required'),
  note: z.string().optional(),
});

export type SendRequest = z.infer<typeof sendRequestSchema>;

export const transactionResponseSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  asset_id: z.number(),
  type: z.string(),
  status: z.string(),
  amount: z.string(),
  fee: z.string().optional(),
  to_address: z.string().optional().nullable(),
  note: z.string().optional().nullable(),
  created_at: z.string(),
});

export type TransactionResponse = z.infer<typeof transactionResponseSchema>;
