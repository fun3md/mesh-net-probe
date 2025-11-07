// Minimal, correct auth context implementation without JSX in this file.
// App.tsx already uses <AuthProvider>, so we expose AuthProvider and useAuth.

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from 'react';
import apiService from '@/services/api';
import type { User } from '@/types';

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user';

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
}

interface AuthContextValue extends AuthState {
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function readInitialState(): AuthState {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    const rawUser = localStorage.getItem(USER_KEY);

    if (token && rawUser) {
      const user = JSON.parse(rawUser) as User;
      apiService.setAuthToken(token);
      return {
        user,
        token,
        isAuthenticated: true,
        loading: false,
      };
    }
  } catch {
    // ignore and fall through
  }

  apiService.clearAuthToken();
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);

  return {
    user: null,
    token: null,
    isAuthenticated: false,
    loading: false,
  };
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: null,
    isAuthenticated: false,
    loading: true,
  });

  // Hydrate once on mount
  useEffect(() => {
    const initial = readInitialState();
    console.debug('[Auth] Hydrated initial auth state from storage', initial);
    setState(initial);
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    console.debug('[Auth] login() called', { username });
    setState(prev => ({ ...prev, loading: true }));

    try {
      const { token, user } = await apiService.login({ username, password });
      console.debug('[Auth] login() success, setting token/user', { user, hasToken: !!token });

      localStorage.setItem(TOKEN_KEY, token);
      localStorage.setItem(USER_KEY, JSON.stringify(user));
      apiService.setAuthToken(token);

      setState({
        user,
        token,
        isAuthenticated: true,
        loading: false,
      });
    } catch (error) {
      console.debug('[Auth] login() failed, clearing auth state', error);
      apiService.clearAuthToken();
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);

      setState({
        user: null,
        token: null,
        isAuthenticated: false,
        loading: false,
      });
      throw error;
    }
  }, []);

  const logout = useCallback(async () => {
    console.debug('[Auth] logout() called');
    setState(prev => ({ ...prev, loading: true }));
    try {
      await apiService.logout().catch(err => {
        console.debug('[Auth] logout() backend error ignored', err);
      });
    } finally {
      console.debug('[Auth] logout() clearing local auth state');
      apiService.clearAuthToken();
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);

      setState({
        user: null,
        token: null,
        isAuthenticated: false,
        loading: false,
      });
    }
  }, []);

  const refresh = useCallback(async () => {
    if (!state.token) {
      console.debug('[Auth] refresh() skipped: no token in state');
      return;
    }

    console.debug('[Auth] refresh() called');
    try {
      const { token } = await apiService.refreshToken();
      console.debug('[Auth] refresh() success, updating token');
      localStorage.setItem(TOKEN_KEY, token);
      apiService.setAuthToken(token);
      setState(prev => ({
        ...prev,
        token,
        isAuthenticated: true,
      }));
    } catch (err) {
      console.debug('[Auth] refresh() failed, performing logout()', err);
      await logout();
    }
  }, [state.token, logout]);

  useEffect(() => {
    console.debug('[Auth] State changed', state);
  }, [state]);

  const value: AuthContextValue = {
    user: state.user,
    token: state.token,
    isAuthenticated: state.isAuthenticated,
    loading: state.loading,
    login,
    logout,
    refresh,
  };

  // Use React.createElement instead of JSX to avoid TSX parsing/encoding issues
  return React.createElement(AuthContext.Provider, { value }, children);
};

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return ctx;
}