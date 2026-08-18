import apiRequest from '@/lib/apiRequest';
import {
  depositAddressResponseSchema,
  transactionResponseSchema,
  type DepositAddressResponse,
  type SendRequest,
  type TransactionResponse,
} from './transfer.schema';

export const transferService = {
  getDepositAddress: async (assetId: number): Promise<DepositAddressResponse> => {
    const response = await apiRequest.get(`/wallet/deposit/${assetId}`);
    return depositAddressResponseSchema.parse(response.data);
  },

  sendCrypto: async (data: SendRequest): Promise<TransactionResponse> => {
    const response = await apiRequest.post('/wallet/send', data);
    return transactionResponseSchema.parse(response.data);
  },
};
