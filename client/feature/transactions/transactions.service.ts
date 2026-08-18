import apiRequest from '@/lib/apiRequest';
import {
  transactionsResponseSchema,
  type TransactionsResponse,
} from './transactions.schema';

export interface GetTransactionsParams {
  page?: number;
  limit?: number;
  type?: string;
}

export const transactionsService = {
  getTransactions: async (params?: GetTransactionsParams): Promise<TransactionsResponse> => {
    const response = await apiRequest.get('/wallet/transactions', {
      params,
    });
    return transactionsResponseSchema.parse(response.data);
  },
};
