import apiRequest from '@/lib/apiRequest';
import { z } from 'zod';
import { AuthUser, authUserSchema } from '../auth/auth.schema';

export const adminUserSchema = authUserSchema.extend({
  phone: z.string().nullable().optional(),
  kyc_status: z.string(),
  created_at: z.string(),
});

export type AdminUser = z.infer<typeof adminUserSchema>;

export const adminService = {
  getUsers: async (): Promise<AdminUser[]> => {
    const response = await apiRequest.get('/admin/users');
    return z.array(adminUserSchema).parse(response.data.users);
  },

  updateUserRole: async (userId: string, role: string): Promise<void> => {
    await apiRequest.put(`/admin/users/${userId}/role`, { role });
  },

  getTransactions: async (): Promise<any[]> => {
    const response = await apiRequest.get('/admin/transactions');
    return response.data.transactions;
  },
};
