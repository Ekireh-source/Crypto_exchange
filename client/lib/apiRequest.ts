import axios from 'axios';
import { store } from '@/store';
import { setTokens, logout } from '@/store/authSlice';

// Create a reusable axios instance
const apiRequest = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to attach the access token
apiRequest.interceptors.request.use(
  (config) => {
    // Get the current token from the Redux store
    const { accessToken } = store.getState().auth;
    
    if (accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle token refresh on 401
apiRequest.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // If the error is 401 and we haven't tried refreshing yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const { refreshToken } = store.getState().auth;

        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        // Call the refresh endpoint
        const response = await axios.post(
          `${apiRequest.defaults.baseURL}/auth/refresh`,
          { refresh_token: refreshToken }
        );

        const newAccessToken = response.data.access_token;
        const newRefreshToken = response.data.refresh_token;

        // Update the tokens in Redux (and localStorage via slice logic)
        store.dispatch(
          setTokens({
            accessToken: newAccessToken,
            refreshToken: newRefreshToken,
          })
        );

        // Update the authorization header for the original request and retry
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
        return apiRequest(originalRequest);
        
      } catch (refreshError) {
        // If refresh fails, log the user out
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
