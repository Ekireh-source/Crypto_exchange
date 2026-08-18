import apiRequest from '@/lib/apiRequest';
import {
  loginSchema,
  registerSchema,
  authResponseSchema,
  type LoginDto,
  type RegisterDto,
  type AuthResponse,
} from './auth.schema';

export type { LoginDto, RegisterDto, AuthResponse };

export const authService = {
  login: async (data: LoginDto): Promise<AuthResponse> => {
    // Validate input before sending
    loginSchema.parse(data);

    const response = await apiRequest.post('/auth/login', data);
    // Parse and validate the API response shape
    return authResponseSchema.parse(response.data);
  },

  register: async (data: RegisterDto): Promise<AuthResponse> => {
    // Validate input before sending
    registerSchema.parse(data);

    const response = await apiRequest.post('/auth/register', data);
    return authResponseSchema.parse(response.data);
  },
};
