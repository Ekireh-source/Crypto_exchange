import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface User {
  id: string;
  email: string;
  referral_code: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
}

// Load initial state from localStorage if available (only on client side)
const loadInitialState = (): AuthState => {
  if (typeof window === 'undefined') {
    return { user: null, accessToken: null, refreshToken: null, isAuthenticated: false };
  }
  
  try {
    const authData = localStorage.getItem('auth_data');
    if (authData) {
      const parsed = JSON.parse(authData);
      return {
        ...parsed,
        isAuthenticated: !!parsed.accessToken,
      };
    }
  } catch (e) {
    console.error('Failed to parse auth data from local storage', e);
  }
  
  return { user: null, accessToken: null, refreshToken: null, isAuthenticated: false };
};

const initialState: AuthState = loadInitialState();

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ user: User; accessToken: string; refreshToken: string }>
    ) => {
      state.user = action.payload.user;
      state.accessToken = action.payload.accessToken;
      state.refreshToken = action.payload.refreshToken;
      state.isAuthenticated = true;

      if (typeof window !== 'undefined') {
        localStorage.setItem(
          'auth_data',
          JSON.stringify({
            user: state.user,
            accessToken: state.accessToken,
            refreshToken: state.refreshToken,
          })
        );
      }
    },
    setTokens: (state, action: PayloadAction<{ accessToken: string; refreshToken: string }>) => {
      state.accessToken = action.payload.accessToken;
      state.refreshToken = action.payload.refreshToken;
      
      if (typeof window !== 'undefined') {
        const currentData = JSON.parse(localStorage.getItem('auth_data') || '{}');
        localStorage.setItem(
          'auth_data',
          JSON.stringify({
            ...currentData,
            accessToken: state.accessToken,
            refreshToken: state.refreshToken,
          })
        );
      }
    },
    logout: (state) => {
      state.user = null;
      state.accessToken = null;
      state.refreshToken = null;
      state.isAuthenticated = false;
      
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_data');
      }
    },
  },
});

export const { setCredentials, setTokens, logout } = authSlice.actions;

export default authSlice.reducer;
