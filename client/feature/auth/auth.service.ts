import apiRequest from '@/lib/apiRequest';

export interface RegisterDto {
  email: string;
  password: string;
  referral_code?: string;
}

export interface LoginDto {
  email: string;
  password: string;
}

export const authService = {
  register: async (data: RegisterDto) => {
    const response = await apiRequest.post('/auth/register', data);
    return response.data;
  },

  login: async (data: LoginDto) => {
    const response = await apiRequest.post('/auth/login', data);
    return response.data;
  },
};
