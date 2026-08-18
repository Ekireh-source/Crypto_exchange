import apiRequest from '@/lib/apiRequest';
import { assetsResponseSchema, type Asset } from './assets.schema';

export const assetsService = {
  getAssets: async (): Promise<Asset[]> => {
    const response = await apiRequest.get('/wallet/assets');
    return assetsResponseSchema.parse(response.data);
  },
};
