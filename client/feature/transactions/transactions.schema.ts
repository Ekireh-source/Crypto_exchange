import { z } from 'zod';

export const transactionItemSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  asset_id: z.number(),
  type: z.string(),
  status: z.string(),
  amount: z.string(),
  fee: z.string().optional().nullable(),
  tx_hash: z.string().optional().nullable(),
  from_address: z.string().optional().nullable(),
  to_address: z.string().optional().nullable(),
  note: z.string().optional().nullable(),
  created_at: z.string(),
  confirmed_at: z.string().optional().nullable(),
});

export type TransactionItem = z.infer<typeof transactionItemSchema>;

export const transactionsResponseSchema = z.object({
  transactions: z.array(transactionItemSchema).nullable().optional(),
  total: z.number().optional().default(0),
  page: z.number().optional().default(1),
  limit: z.number().optional().default(20),
});

export type TransactionsResponse = z.infer<typeof transactionsResponseSchema>;
