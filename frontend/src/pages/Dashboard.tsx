import { useEffect, useState } from 'react';
import { useQuery, useMutation, useSubscription, gql } from '@apollo/client';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import type { RootState } from '../store';
import { logout } from '../store/authSlice';
import { setBalance } from '../store/balanceSlice';
import toast from 'react-hot-toast';

// ===== ЗАПРОСЫ =====
const GET_BALANCE = gql`
  query GetBalance($accountId: ID!) {
    balance(accountID: $accountId) {
      id
      balance
      currency
    }
  }
`;

const TRANSFER_MUTATION = gql`
  mutation Transfer($toAccountId: String!, $amount: Float!, $currency: String!, $description: String) {
    transfer(toAccountID: $toAccountId, amount: $amount, currency: $currency, description: $description) {
      id
      fromAccountID
      toAccountID
      amount
      currency
      status
      description
      createdAt
    }
  }
`;

const GET_TRANSACTIONS = gql`
  query GetTransactions($limit: Int, $offset: Int) {
    transactions(limit: $limit, offset: $offset) {
      id
      fromAccountID
      toAccountID
      amount
      currency
      status
      description
      createdAt
    }
  }
`;

const TRANSACTION_SUBSCRIPTION = gql`
  subscription OnTransactionCompleted {
    transactionCompleted {
      id
      fromAccountID
      toAccountID
      amount
      currency
      status
      description
      createdAt
    }
  }
`;

export default function Dashboard() {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { user, accessToken } = useSelector((state: RootState) => state.auth);
  const { balance, currency } = useSelector((state: RootState) => state.balance);

  // ===== РЕАЛЬНЫЙ ACCOUNT ID =====
  // ВСТАВЬ СЮДА ID СЧЁТА ИЗ БД
  const REAL_ACCOUNT_ID = '6011a9d7-2e44-4980-a11c-c8a3cc1f6067';

  // ===== БАЛАНС =====
  const {
    data: balanceData,
    loading: balanceLoading,
    refetch: refetchBalance,
  } = useQuery(GET_BALANCE, {
    variables: { accountId: REAL_ACCOUNT_ID },
    fetchPolicy: 'network-only', // всегда свежие данные
    skip: !REAL_ACCOUNT_ID,
  });



  useEffect(() => {
    if (balanceData) {
      console.log('✅ Balance loaded:', balanceData.balance);
      dispatch(setBalance({
        balance: balanceData.balance.balance,
        currency: balanceData.balance.currency,
        accountId: balanceData.balance.id,
      }));
    }
  }, [balanceData, dispatch]);

  // ===== ТРАНЗАКЦИИ =====
  const {
    data: txData,
    loading: txLoading,
    refetch: refetchTransactions,
  } = useQuery(GET_TRANSACTIONS, {
    variables: { limit: 10, offset: 0 },
    fetchPolicy: 'network-only',
  });

  // ===== ПОДПИСКА =====
  const { data: subData } = useSubscription(TRANSACTION_SUBSCRIPTION, {
    skip: !accessToken,
  });

  useEffect(() => {
    if (subData?.transactionCompleted) {
      console.log('🔄 New transaction via WS, refreshing...');
      toast.success('New transaction received!');
      refetchBalance();
      refetchTransactions();
    }
  }, [subData, refetchBalance, refetchTransactions]);

  // ===== ЛОГАУТ =====
  const handleLogout = () => {
    dispatch(logout());
    navigate('/login');
  };

  // ===== ФОРМА ПЕРЕВОДА =====
  const [toAccount, setToAccount] = useState('');
  const [amount, setAmount] = useState('');
  const [currencyField, setCurrencyField] = useState('USD');

  const [transfer, { loading: transferLoading }] = useMutation(TRANSFER_MUTATION, {
    onCompleted: (data) => {
      console.log('✅ Transfer completed:', data);
      toast.success('Transfer completed!');
      console.log('🔁 Refetching balance...');
      refetchBalance();
      console.log('🔁 Refetching transactions...');
      refetchTransactions();
    },
    onError: (error) => {
      console.error('❌ Transfer error:', error);
    },
  });



  const handleTransfer = (e: React.FormEvent) => {
    e.preventDefault();
    if (!toAccount || !amount) {
      toast.error('Please fill in all fields');
      return;
    }
    console.log('📤 Sending transfer:', { toAccount, amount, currencyField });
    transfer({
      variables: {
        toAccountId: toAccount,
        amount: parseFloat(amount),
        currency: currencyField,
        description: 'Transfer from dashboard',
      },
    });
  };

  // ===== UI =====
  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)',
      padding: '32px 16px',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
          <h1 style={{ fontSize: '28px', fontWeight: 'bold', color: '#1e293b' }}>Dashboard</h1>
          <button
            onClick={handleLogout}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              background: 'white',
              padding: '10px 20px',
              borderRadius: '12px',
              border: '1px solid #e2e8f0',
              boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
              cursor: 'pointer',
              transition: 'all 0.2s',
              fontSize: '14px',
              fontWeight: '500',
              color: '#334155',
            }}
            onMouseEnter={(e) => (e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.08)')}
            onMouseLeave={(e) => (e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.06)')}
          >
            <span>🚪</span> Logout
          </button>
        </div>

        {/* Balance Card */}
        <div style={{
          background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
          borderRadius: '16px',
          padding: '24px 32px',
          marginBottom: '32px',
          color: 'white',
          boxShadow: '0 10px 40px rgba(99, 102, 241, 0.3)',
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
            <div>
              <p style={{ color: 'rgba(255,255,255,0.8)', fontSize: '14px', margin: 0 }}>Total Balance</p>
              {balanceLoading ? (
                <div style={{ width: '120px', height: '40px', background: 'rgba(255,255,255,0.2)', borderRadius: '8px', marginTop: '4px', animation: 'pulse 1.5s ease-in-out infinite' }}></div>
              ) : (
                <p style={{ fontSize: '42px', fontWeight: 'bold', margin: '4px 0 0 0' }}>
                  {balance} <span style={{ fontSize: '20px', fontWeight: '400' }}>{currency}</span>
                </p>
              )}
            </div>
            <div style={{ background: 'rgba(255,255,255,0.15)', padding: '8px 16px', borderRadius: '12px', backdropFilter: 'blur(4px)' }}>
              <span style={{ fontSize: '14px' }}>👤 {user?.fullName || 'User'}</span>
            </div>
          </div>
        </div>

        {/* Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '24px' }}>
          {/* Transfer Form */}
          <div style={{
            background: 'white',
            borderRadius: '16px',
            padding: '24px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
            border: '1px solid #e2e8f0',
          }}>
            <h3 style={{ fontSize: '18px', fontWeight: '600', color: '#1e293b', margin: '0 0 16px 0' }}>
              💸 Transfer Money
            </h3>
            <form onSubmit={handleTransfer} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              <div>
                <label style={{ fontSize: '12px', fontWeight: '500', color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                  Recipient Account
                </label>
                <input
                  type="text"
                  placeholder="Account ID"
                  value={toAccount}
                  onChange={(e) => setToAccount(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '12px 16px',
                    border: '1px solid #e2e8f0',
                    borderRadius: '12px',
                    fontSize: '14px',
                    outline: 'none',
                    transition: 'border-color 0.2s',
                    boxSizing: 'border-box',
                    marginTop: '4px',
                  }}
                  onFocus={(e) => (e.currentTarget.style.borderColor = '#6366f1')}
                  onBlur={(e) => (e.currentTarget.style.borderColor = '#e2e8f0')}
                  required
                />
              </div>
              <div>
                <label style={{ fontSize: '12px', fontWeight: '500', color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                  Amount
                </label>
                <input
                  type="number"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '12px 16px',
                    border: '1px solid #e2e8f0',
                    borderRadius: '12px',
                    fontSize: '14px',
                    outline: 'none',
                    transition: 'border-color 0.2s',
                    boxSizing: 'border-box',
                    marginTop: '4px',
                  }}
                  onFocus={(e) => (e.currentTarget.style.borderColor = '#6366f1')}
                  onBlur={(e) => (e.currentTarget.style.borderColor = '#e2e8f0')}
                  required
                />
              </div>
              <div>
                <label style={{ fontSize: '12px', fontWeight: '500', color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                  Currency
                </label>
                <select
                  value={currencyField}
                  onChange={(e) => setCurrencyField(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '12px 16px',
                    border: '1px solid #e2e8f0',
                    borderRadius: '12px',
                    fontSize: '14px',
                    outline: 'none',
                    background: 'white',
                    boxSizing: 'border-box',
                    marginTop: '4px',
                  }}
                >
                  <option value="USD">🇺🇸 USD</option>
                  <option value="EUR">🇪🇺 EUR</option>
                  <option value="RUB">🇷🇺 RUB</option>
                </select>
              </div>
              <button
                type="submit"
                disabled={transferLoading}
                style={{
                  width: '100%',
                  padding: '14px',
                  background: transferLoading ? '#94a3b8' : '#6366f1',
                  color: 'white',
                  border: 'none',
                  borderRadius: '12px',
                  fontSize: '16px',
                  fontWeight: '600',
                  cursor: transferLoading ? 'not-allowed' : 'pointer',
                  transition: 'background 0.2s',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                }}
                onMouseEnter={(e) => {
                  if (!transferLoading) e.currentTarget.style.background = '#4f46e5';
                }}
                onMouseLeave={(e) => {
                  if (!transferLoading) e.currentTarget.style.background = '#6366f1';
                }}
              >
                {transferLoading ? '⏳ Processing...' : '🚀 Send Money'}
              </button>
            </form>
          </div>

          {/* Transactions History */}
          <div style={{
            background: 'white',
            borderRadius: '16px',
            padding: '24px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
            border: '1px solid #e2e8f0',
          }}>
            <h3 style={{ fontSize: '18px', fontWeight: '600', color: '#1e293b', margin: '0 0 16px 0' }}>
              📋 Recent Transactions
            </h3>
            <TransactionsList data={txData} loading={txLoading} refetch={refetchTransactions} />
          </div>
        </div>
      </div>

      <style>{`
        @keyframes pulse {
          0% { opacity: 0.6; }
          50% { opacity: 1; }
          100% { opacity: 0.6; }
        }
      `}</style>
    </div>
  );
}

// ----- Компонент списка транзакций (упрощённый, с refetch) -----
function TransactionsList({ data, loading, refetch }: any) {
  const [page, setPage] = useState(1);
  const [limit] = useState(5);
  const offset = (page - 1) * limit;
  const transactions = data?.transactions || [];
  const filtered = transactions;
  const totalPages = Math.ceil(transactions.length / limit) || 1;

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '32px 0' }}>
        <div style={{ display: 'inline-block', width: '32px', height: '32px', border: '3px solid #e2e8f0', borderTopColor: '#6366f1', borderRadius: '50%', animation: 'spin 0.8s linear infinite' }}></div>
        <style>{`
          @keyframes spin {
            to { transform: rotate(360deg); }
          }
        `}</style>
      </div>
    );
  }

  if (!transactions.length) {
    return <p style={{ textAlign: 'center', color: '#94a3b8', padding: '24px 0' }}>No transactions yet</p>;
  }

  const paginated = filtered.slice(offset, offset + limit);

  return (
    <div>
      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', fontSize: '14px', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
              <th style={{ textAlign: 'left', padding: '12px 8px', fontWeight: '600', color: '#64748b' }}>Amount</th>
              <th style={{ textAlign: 'left', padding: '12px 8px', fontWeight: '600', color: '#64748b' }}>Currency</th>
              <th style={{ textAlign: 'left', padding: '12px 8px', fontWeight: '600', color: '#64748b' }}>Status</th>
              <th style={{ textAlign: 'left', padding: '12px 8px', fontWeight: '600', color: '#64748b', display: 'none' }}>Description</th>
              <th style={{ textAlign: 'left', padding: '12px 8px', fontWeight: '600', color: '#64748b' }}>Date</th>
            </tr>
          </thead>
          <tbody>
            {paginated.map((tx: any) => (
              <tr key={tx.id} style={{ borderBottom: '1px solid #f1f5f9' }}>
                <td style={{ padding: '12px 8px', fontWeight: '500', color: '#1e293b' }}>{tx.amount}</td>
                <td style={{ padding: '12px 8px', color: '#475569' }}>{tx.currency}</td>
                <td style={{ padding: '12px 8px' }}>
                  <span style={{
                    display: 'inline-block',
                    padding: '2px 10px',
                    borderRadius: '9999px',
                    fontSize: '11px',
                    fontWeight: '600',
                    background: tx.status === 'completed' ? '#d1fae5' : tx.status === 'pending' ? '#fef3c7' : '#fee2e2',
                    color: tx.status === 'completed' ? '#065f46' : tx.status === 'pending' ? '#92400e' : '#991b1b',
                  }}>
                    {tx.status}
                  </span>
                </td>
                <td style={{ padding: '12px 8px', color: '#64748b', display: 'none' }}>{tx.description || '—'}</td>
                <td style={{ padding: '12px 8px', color: '#94a3b8', fontSize: '12px' }}>{new Date(tx.createdAt).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '16px', flexWrap: 'wrap', gap: '8px' }}>
        <span style={{ fontSize: '13px', color: '#94a3b8' }}>Page {page} of {totalPages}</span>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            onClick={() => setPage(p => Math.max(p - 1, 1))}
            disabled={page === 1}
            style={{
              padding: '6px 16px',
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              background: 'white',
              fontSize: '13px',
              fontWeight: '500',
              color: '#334155',
              cursor: page === 1 ? 'not-allowed' : 'pointer',
              opacity: page === 1 ? 0.4 : 1,
            }}
          >
            ← Previous
          </button>
          <button
            onClick={() => setPage(p => Math.min(p + 1, totalPages))}
            disabled={page === totalPages}
            style={{
              padding: '6px 16px',
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              background: 'white',
              fontSize: '13px',
              fontWeight: '500',
              color: '#334155',
              cursor: page === totalPages ? 'not-allowed' : 'pointer',
              opacity: page === totalPages ? 0.4 : 1,
            }}
          >
            Next →
          </button>
        </div>
      </div>
    </div>
  );
}