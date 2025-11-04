import { useState, useEffect } from 'react';
import { apiService } from '@/services/api';
import type { User, LoginRequest } from '@/types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
}

interface AuthActions {
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
}

export const useAuth = (): AuthState & AuthActions => {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    loading: true,
  });

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      if (!token) {
        setState(prev => ({ ...prev, loading: false }));
        return;
      }

      const user = await apiService.getCurrentUser();
      setState({
        user,
        isAuthenticated: true,
        loading: false,
      });
    } catch (error) {
      // Token might be invalid
      console.error('Auth check failed:', error);
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user');
      setState({
        user: null,
        isAuthenticated: false,
        loading: false,
      });
    }
  };

  const login = async (credentials: LoginRequest) => {
    try {
      setState(prev => ({ ...prev, loading: true }));
      
      const response = await apiService.login(credentials);
      
      // Store tokens and user data
      // Backend returns: { token, user, expiresAt }
      apiService.setAuthToken(response.token);
      localStorage.setItem('user', JSON.stringify(response.user));
      
      setState({
        user: response.user,
        isAuthenticated: true,
        loading: false,
      });
    } catch (error) {
      setState(prev => ({ ...prev, loading: false }));
      throw error;
    }
  };

  const logout = async () => {
    try {
      await apiService.logout();
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      setState({
        user: null,
        isAuthenticated: false,
        loading: false,
      });
    }
  };

  const refreshToken = async () => {
    try {
      const newToken = await apiService.refreshToken();
      apiService.setAuthToken(newToken.accessToken);
    } catch (error) {
      console.error('Token refresh failed:', error);
      // If refresh fails, redirect to login
      await logout();
    }
  };

  return {
    ...state,
    login,
    logout,
    refreshToken,
  };
};