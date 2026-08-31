import { configureStore, combineReducers, AnyAction } from '@reduxjs/toolkit';
import authReducer from './authSlice';

const appReducer = combineReducers({
  auth: authReducer,
});

const rootReducer = (state: any, action: AnyAction) => {
  if (action.type === 'auth/logout') {
    // Reset the entire state to undefined so that reducers revert to initial state
    state = undefined;
  }
  return appReducer(state, action);
};

export const store = configureStore({
  reducer: rootReducer,
});

export type RootState = ReturnType<typeof appReducer>;
export type AppDispatch = typeof store.dispatch;
