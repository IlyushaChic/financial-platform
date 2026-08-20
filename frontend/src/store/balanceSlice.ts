import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';

interface BalanceState {
  balance: number;
  currency: string;
  accountId: string | null;
}

const initialState: BalanceState = {
  balance: 0,
  currency: 'USD',
  accountId: null,
};

const balanceSlice = createSlice({
  name: 'balance',
  initialState,
  reducers: {
    setBalance: (state, action: PayloadAction<{ balance: number; currency: string; accountId: string }>) => {
      state.balance = action.payload.balance;
      state.currency = action.payload.currency;
      state.accountId = action.payload.accountId;
    },
    updateBalance: (state, action: PayloadAction<number>) => {
      state.balance = action.payload;
    },
  },
});

export const { setBalance, updateBalance } = balanceSlice.actions;
export default balanceSlice.reducer;
