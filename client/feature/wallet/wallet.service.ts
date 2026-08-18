import apiRequest from '@/lib/apiRequest';
import {
  portfolioResponseSchema,
  type PortfolioResponse,
} from './wallet.schema';

export const walletService = {
  getPortfolio: async (): Promise<PortfolioResponse> => {
    const response = await apiRequest.get('/wallet/portfolio');
    return portfolioResponseSchema.parse(response.data);
  },
};
