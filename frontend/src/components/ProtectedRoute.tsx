import { Navigate, Outlet } from 'react-router-dom';
import { useSelector } from 'react-redux';
import type { RootState } from '../store';

export const ProtectedRoute = () => {
  const accessToken = useSelector((state: RootState) => state.auth.accessToken);
  return accessToken ? <Outlet /> : <Navigate to="/login" replace />;
};