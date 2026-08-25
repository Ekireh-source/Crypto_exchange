
import apiRequest from '@/lib/apiRequest';
import { type UserProfileResponse, userProfileResponseSchema } from './user.schema';

export const userService = {
  getProfile: async (): Promise<UserProfileResponse> => {
    const response = await apiRequest.get('/user/profile');
    return userProfileResponseSchema.parse(response.data);
  },
};
