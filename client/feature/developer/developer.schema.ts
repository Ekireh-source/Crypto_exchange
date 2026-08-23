import { z } from 'zod';

// ── Application ───────────────────────────────────────────────────────────────

export const apiApplicationSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  name: z.string(),
  is_live: z.boolean(),
  created_at: z.string(),
});
export type APIApplication = z.infer<typeof apiApplicationSchema>;

// ── Key Pair & Generation ─────────────────────────────────────────────────────

export const generatedKeyPairSchema = z.object({
  secret_key: z.string(),
  publishable_key: z.string(),
});
export type GeneratedKeyPair = z.infer<typeof generatedKeyPairSchema>;

export const createAppResponseSchema = z.object({
  app: apiApplicationSchema,
  keys: generatedKeyPairSchema,
  message: z.string(),
});
export type CreateAppResponse = z.infer<typeof createAppResponseSchema>;

// ── API Key ───────────────────────────────────────────────────────────────────

export const apiKeySchema = z.object({
  id: z.string(),
  app_id: z.string(),
  secret_hint: z.string(),
  publishable_hint: z.string(),
  is_active: z.boolean(),
  last_used_at: z.string().nullish(),
  created_at: z.string(),
});
export type APIKey = z.infer<typeof apiKeySchema>;

export const regenerateKeysResponseSchema = z.object({
  keys: generatedKeyPairSchema,
  message: z.string(),
});
export type RegenerateKeysResponse = z.infer<typeof regenerateKeysResponseSchema>;

// ── Logs ──────────────────────────────────────────────────────────────────────

export const requestLogSchema = z.object({
  id: z.number(),
  key_id: z.string(),
  method: z.string(),
  path: z.string(),
  status_code: z.number(),
  latency_ms: z.number(),
  created_at: z.string(),
});
export type RequestLog = z.infer<typeof requestLogSchema>;

export const getLogsResponseSchema = z.object({
  logs: z.array(requestLogSchema).nullable().transform(val => val ?? []),
  total: z.number(),
  page: z.number(),
  limit: z.number(),
});
export type GetLogsResponse = z.infer<typeof getLogsResponseSchema>;

// ── Webhooks ──────────────────────────────────────────────────────────────────

export const webhookSchema = z.object({
  id: z.string(),
  app_id: z.string(),
  url: z.string(),
  events: z.array(z.string()),
  secret: z.string().optional(),
  is_active: z.boolean(),
  created_at: z.string(),
});
export type Webhook = z.infer<typeof webhookSchema>;

export const createWebhookResponseSchema = z.object({
  webhook: webhookSchema,
  message: z.string(),
});
export type CreateWebhookResponse = z.infer<typeof createWebhookResponseSchema>;
