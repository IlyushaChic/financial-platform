import { useState } from 'react';
import { gql, useMutation } from '@apollo/client';
import { useDispatch } from 'react-redux';
import { useNavigate, Link } from 'react-router-dom';
import { setCredentials } from '../store/authSlice';
import toast from 'react-hot-toast';

interface RegisterResponse {
  register: {
    user: { id: string };
    message: string;
  };
}

interface RegisterVariables {
  email: string;
  password: string;
  fullName: string;
}

interface LoginResponse {
  login: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    user: {
      id: string;
      email: string;
      fullName: string;
    };
  };
}

interface LoginVariables {
  email: string;
  password: string;
}

const REGISTER_MUTATION = gql`
  mutation Register($email: String!, $password: String!, $fullName: String!) {
    register(email: $email, password: $password, fullName: $fullName) {
      user {
        id
      }
      message
    }
  }
`;

const LOGIN_MUTATION = gql`
  mutation Login($email: String!, $password: String!) {
    login(email: $email, password: $password) {
      accessToken
      refreshToken
      expiresIn
      user {
        id
        email
        fullName
      }
    }
  }
`;

export default function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fullName, setFullName] = useState('');
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const [register, { loading: registerLoading }] = useMutation<RegisterResponse, RegisterVariables>(REGISTER_MUTATION);

  const [login, { loading: loginLoading }] = useMutation<LoginResponse, LoginVariables>(LOGIN_MUTATION, {
    onCompleted: (data) => {
      const { user, accessToken, refreshToken } = data.login;
      dispatch(setCredentials({ user, accessToken, refreshToken }));
      toast.success('Registered and logged in!');
      navigate('/dashboard');
    },
    onError: (error) => {
      toast.error(error.message || 'Registration failed');
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await register({ variables: { email, password, fullName } });
      await login({ variables: { email, password } });
    } catch {
      // ошибка уже обработана в onError логина
    }
  };

  const loading = registerLoading || loginLoading;

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8">
        <h2 className="text-3xl font-bold text-center">Create account</h2>
        <form onSubmit={handleSubmit} className="space-y-6">
          <input
            type="text"
            placeholder="Full Name"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            className="w-full p-2 border rounded"
            required
          />
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full p-2 border rounded"
            required
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full p-2 border rounded"
            required
          />
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-green-600 text-white p-2 rounded hover:bg-green-700 disabled:opacity-50"
          >
            {loading ? 'Loading...' : 'Register'}
          </button>
        </form>
        <p className="text-center">
          Already have an account? <Link to="/login" className="text-blue-600">Sign in</Link>
        </p>
      </div>
    </div>
  );
}