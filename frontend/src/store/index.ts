import { configureStore } from '@reduxjs/toolkit';
import authReducer from './authSlice';
import balanceReducer from './balanceSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    balance: balanceReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;