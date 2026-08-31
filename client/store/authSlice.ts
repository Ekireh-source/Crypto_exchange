import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { RootState } from './index';

interface User {
  id: string;
  email: string;
  role: string;
  referral_code: string;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
}

const loadInitialState = (): AuthState => {
  if (typeof window === 'undefined') {
    return { user: null, isAuthenticated: false };
  }
  
  try {
    const authData = localStorage.getItem('auth_data');
    if (authData) {
      const parsed = JSON.parse(authData);
      return {
        user: parsed.user,
        isAuthenticated: !!parsed.user,
      };
    }
  } catch (e) {
    console.error('Failed to parse auth data from local storage', e);
  }
  
  return { user: null, isAuthenticated: false };
};

const initialState: AuthState = loadInitialState();

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ user: User }>
    ) => {
      state.user = action.payload.user;
      state.isAuthenticated = true;

      if (typeof window !== 'undefined') {
        localStorage.setItem(
          'auth_data',
          JSON.stringify({
            user: state.user,
          })
        );
      }
    },
    logout: (state) => {
      state.user = null;
      state.isAuthenticated = false;
      
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_data');
      }
    },
  },
});

export const { setCredentials, logout } = authSlice.actions;

export const selectUser = (state: RootState) => state.auth.user;

export default authSlice.reducer;
