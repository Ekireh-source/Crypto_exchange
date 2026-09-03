import axios from 'axios';
import { store } from '@/store';
import { logout } from '@/store/authSlice';

// Create a reusable axios instance
const apiRequest = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

// Request interceptor to add Idempotency-Key for mutating requests
apiRequest.interceptors.request.use((config) => {
  const mutatingMethods = ['post', 'put', 'patch', 'delete'];
  if (config.method && mutatingMethods.includes(config.method.toLowerCase())) {
    if (!config.headers['Idempotency-Key']) {
      // crypto.randomUUID() is available in secure contexts (HTTPS or localhost)
      if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        config.headers['Idempotency-Key'] = crypto.randomUUID();
      } else {
        // Fallback for older environments
        config.headers['Idempotency-Key'] = Math.random().toString(36).substring(2) + Date.now().toString(36);
      }
    }
  }
  return config;
});

// Response interceptor to handle token refresh on 401
apiRequest.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // If the error is 401 and we haven't tried refreshing yet
    const isAuthRoute = originalRequest.url?.includes('/auth/login') || 
                        originalRequest.url?.includes('/auth/register') || 
                        originalRequest.url?.includes('/auth/refresh');

    if (error.response?.status === 401 && !originalRequest._retry && !isAuthRoute) {
      originalRequest._retry = true;

      try {
        // Call the refresh endpoint. 
        // We don't need to pass the refresh_token in the body because the browser sends the HttpOnly cookie.
        await axios.post(
          `${apiRequest.defaults.baseURL}/auth/refresh`,
          {},
          { withCredentials: true }
        );

        // The backend automatically sets the new HttpOnly cookies on the response.
        // We can just retry the original request.
        return apiRequest(originalRequest);
        
      } catch (refreshError) {
        // If refresh fails, try to log the user out on the backend to clear cookies
        try {
          await axios.post(`${apiRequest.defaults.baseURL}/auth/logout`, {}, { withCredentials: true });
        } catch (e) {
          // ignore
        }

        // Log out on the frontend
        store.dispatch(logout());
        
        // Optionally redirect to login page (can also be handled in components via store listener)
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default apiRequest;
