import apiRequest from '@/lib/apiRequest';
import {
  portfolioResponseSchema,
  type PortfolioResponse,
  type SwapRequest,
  type SwapResponse,
} from './wallet.schema';

export const walletService = {
  getPortfolio: async (): Promise<PortfolioResponse> => {
    const response = await apiRequest.get('/wallet/portfolio');
    return portfolioResponseSchema.parse(response.data);
  },
  swap: async (data: SwapRequest): Promise<SwapResponse> => {
    const response = await apiRequest.post('/wallet/swap', data);
    return response.data; // You can add zod parsing here if needed
  },
  getWatchlist: async (): Promise<number[]> => {
    const response = await apiRequest.get('/wallet/watchlist');
    return response.data;
  },
  toggleWatchlist: async (assetId: number): Promise<void> => {
    await apiRequest.post('/wallet/watchlist/toggle', { asset_id: assetId });
  },
};
