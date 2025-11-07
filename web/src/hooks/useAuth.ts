import { useState, useEffect, useCallback } from 'react';
import apiService from '@/services/api';
import type { User } from '@/types';

// Keys for localStorage
const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user';

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
}

let initialAuthCheckStarted = false;

/**
 * useAuth
 *
 * Phase 5.2 alignment:
 * - On mount, if auth_token exists, call /auth/me to validate.
 * - On 401 or network errors, clear auth state and storage.
 * - Exposes loading flag so App routing can avoid unauthenticated flicker.
 */
export function useAuth(): {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
} {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: null,
    isAuthenticated: false,
    loading: true,
  });

  // Internal helpers

  const loadFromStorage = useCallback(() => {
    const token = localStorage.getItem(TOKEN_KEY);
    const rawUser = localStorage.getItem(USER_KEY);

    if (!token) {
      return {
        token: null,
        user: null,
        isAuthenticated: false,
      };
    }

    try {
      const user: User | null = rawUser ? JSON.parse(rawUser) : null;
      return {
        token,
        user,
        isAuthenticated: !!user,
      };
    } catch {
      localStorage.removeItem(USER_KEY);
      return {
        token,
        user: null,
        isAuthenticated: false,
      };
    }
  }, []);

  const applyAuthState = useCallback((next: Partial<AuthState>) => {
    setState(prev => ({
      ...prev,
      ...next,
    }));
  }, []);

  const clearAuth = useCallback(async () => {
    try {
      await apiService.logout().catch(() => {
        // Ignore backend logout errors in demo; focus on local cleanup
      });
    } finally {
      apiService.clearAuthToken();
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      applyAuthState({
        user: null,
        token: null,
        isAuthenticated: false,
      });
    }
  }, [applyAuthState]);

  // Initial auth check on mount
  useEffect(() => {
    let isMounted = true;

    const bootstrap = async () => {
      // Avoid duplicate bootstrap if multiple components use this hook
      if (initialAuthCheckStarted) {
        // Wait one tick allowing other instance to update global storage-backed state
        applyAuthState({ loading: false });
        return;
      }
      initialAuthCheckStarted = true;

      const { token, user, isAuthenticated } = loadFromStorage();

      if (!token) {
        apiService.clearAuthToken();
        if (!isMounted) return;
        applyAuthState({
          user: null,
          token: null,
          isAuthenticated: false,
          loading: false,
        });
        return;
      }

      // Set token on apiService so /auth/me uses it
      apiService.setAuthToken(token);

      try {
        // Validate token and fetch current user from backend
        const me = await apiService.getCurrentUser();
        if (!isMounted) return;

        localStorage.setItem(USER_KEY, JSON.stringify(me));
        applyAuthState({
          user: me,
          token,
          isAuthenticated: true,
          loading: false,
        });
      } catch (error: any) {
        // On 401 or any failure, treat as unauthenticated
        if (!isMounted) return;
        apiService.clearAuthToken();
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
        applyAuthState({
          user: null,
          token: null,
          isAuthenticated: false,
          loading: false,
        });
      }
    };

    bootstrap();

    return () => {
      isMounted = false;
    };
  }, [applyAuthState, loadFromStorage]);

  // Public API

  const login = useCallback(
    async (username: string, password: string) => {
      applyAuthState({ loading: true });

      try {
        // Backend: POST /auth/login → { token, user, expiresAt }
        const { token, user } = await apiService.login({ username, password });

        // Persist
        localStorage.setItem(TOKEN_KEY, token);
        localStorage.setItem(USER_KEY, JSON.stringify(user));
        apiService.setAuthToken(token);

        applyAuthState({
          user,
          token,
          isAuthenticated: true,
          loading: false,
        });
      } catch (error) {
        // On failure, clear any stale state
        apiService.clearAuthToken();
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
        applyAuthState({
          user: null,
          token: null,
          isAuthenticated: false,
          loading: false,
        });
        throw error;
      }
    },
    [applyAuthState]
  );

  const logout = useCallback(async () => {
    applyAuthState({ loading: true });
    await clearAuth();
    applyAuthState({ loading: false });
  }, [applyAuthState, clearAuth])

  const refresh = useCallback(async () => {
    if (!state.token) {
      return;
    }

    try {
      const { token } = await apiService.refreshToken();
      // Update stored token
      localStorage.setItem(TOKEN_KEY, token);
      apiService.setAuthToken(token);
      applyAuthState({ token, isAuthenticated: true });
    } catch {
      // On refresh failure, clear auth
      await clearAuth();
    }
  }, [state.token, applyAuthState, clearAuth]);

  return {
    user: state.user,
    token: state.token,
    isAuthenticated: state.isAuthenticated,
    loading: state.loading,
    login,
    logout,
    refresh,
  };
}