import { useCallback, useMemo, useState } from 'react';
import authService from '../services/authService';
import { setAuthToken } from '../services/apiClient';
import AuthContext from './authContext';

export const AuthProvider = ({ children }) => {
  const [session, setSession] = useState(() => {
    const storedSession = authService.getStoredSession();
    if (storedSession?.token) {
      setAuthToken(storedSession.token);
    }

    return storedSession;
  });

  const applySession = useCallback((nextSession) => {
    setSession(nextSession);
    if (nextSession?.token) {
      setAuthToken(nextSession.token);
    } else {
      setAuthToken(null);
    }
  }, []);

  const login = useCallback(async (credentials) => {
    const nextSession = await authService.login(credentials);
    applySession(nextSession);
    return nextSession;
  }, [applySession]);

  const register = useCallback(async (userData) => {
    const nextSession = await authService.register(userData);
    applySession(nextSession);
    return nextSession;
  }, [applySession]);

  const logout = useCallback(() => {
    authService.clearSession();
    applySession(null);
  }, [applySession]);

  const value = useMemo(() => ({
    user: session?.user || null,
    token: session?.token || null,
    isAuthenticated: Boolean(session?.token),
    login,
    register,
    logout,
  }), [login, logout, register, session]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};
