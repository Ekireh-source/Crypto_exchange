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
  getTransactionById: async (id: string): Promise<any> => {
    // For now returning any, but you can import TransactionItem and parse it
    const response = await apiRequest.get(`/wallet/transactions/${id}`);
    return response.data;
  },
};
