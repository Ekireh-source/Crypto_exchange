import apiRequest from '@/lib/apiRequest';
import { z } from 'zod';
import {
  type APIApplication,
  type APIKey,
  type GeneratedKeyPair,
  type CreateAppResponse,
  type RequestLog,
  type Webhook,
  type GetLogsResponse,
  type RegenerateKeysResponse,
  type CreateWebhookResponse,
  apiApplicationSchema,
  apiKeySchema,
  createAppResponseSchema,
  getLogsResponseSchema,
  webhookSchema,
  regenerateKeysResponseSchema,
  createWebhookResponseSchema,
} from './developer.schema';

export type {
  APIApplication,
  APIKey,
  GeneratedKeyPair,
  CreateAppResponse,
  RequestLog,
  Webhook,
  GetLogsResponse,
};

export const developerService = {
  // Applications
  createApp: async (name: string): Promise<CreateAppResponse> => {
    const response = await apiRequest.post('/developer/apps', { name });
    return createAppResponseSchema.parse(response.data);
  },

  listApps: async (): Promise<APIApplication[]> => {
    const response = await apiRequest.get('/developer/apps');
    return z.array(apiApplicationSchema).parse(response.data || []);
  },

  deleteApp: async (appId: string): Promise<void> => {
    await apiRequest.delete(`/developer/apps/${appId}`);
  },

  updateAppStatus: async (appId: string, isLive: boolean): Promise<void> => {
    await apiRequest.put(`/developer/apps/${appId}/status`, { is_live: isLive });
  },

  // Keys
  listKeys: async (appId: string): Promise<APIKey[]> => {
    const response = await apiRequest.get(`/developer/apps/${appId}/keys`);
    return z.array(apiKeySchema).parse(response.data || []);
  },

  regenerateKeys: async (appId: string): Promise<RegenerateKeysResponse> => {
    const response = await apiRequest.post(`/developer/apps/${appId}/keys/regenerate`);
    return regenerateKeysResponseSchema.parse(response.data);
  },

  revokeKey: async (appId: string, keyId: string): Promise<void> => {
    await apiRequest.delete(`/developer/apps/${appId}/keys/${keyId}`);
  },

  // Logs
  getLogs: async (appId: string, page = 1, limit = 50): Promise<GetLogsResponse> => {
    const response = await apiRequest.get(`/developer/apps/${appId}/logs`, { params: { page, limit } });
    return getLogsResponseSchema.parse(response.data);
  },

  // Webhooks
  createWebhook: async (appId: string, url: string, events: string[]): Promise<CreateWebhookResponse> => {
    const response = await apiRequest.post(`/developer/apps/${appId}/webhooks`, { url, events });
    return createWebhookResponseSchema.parse(response.data);
  },

  listWebhooks: async (appId: string): Promise<Webhook[]> => {
    const response = await apiRequest.get(`/developer/apps/${appId}/webhooks`);
    return z.array(webhookSchema).parse(response.data || []);
  },

  deleteWebhook: async (appId: string, webhookId: string): Promise<void> => {
    await apiRequest.delete(`/developer/apps/${appId}/webhooks/${webhookId}`);
  },
};
