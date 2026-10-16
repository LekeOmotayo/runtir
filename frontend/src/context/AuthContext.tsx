import React, { createContext, useContext, useState, useEffect } from 'react';
import { User, Organisation } from '../types';
import { api } from '../api/client';

interface AuthContextType {
  user: User | null;
  activeOrg: Organisation | null;
  organisations: Organisation[];
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string, orgID?: string) => Promise<void>;
  register: (email: string, password: string, fullName: string, orgName?: string) => Promise<void>;
  logout: () => Promise<void>;
  switchOrg: (orgID: string) => Promise<void>;
  refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [activeOrg, setActiveOrg] = useState<Organisation | null>(null);
  const [organisations, setOrganisations] = useState<Organisation[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const refreshProfile = async () => {
    try {
      const resp = await api.auth.me();
      if (resp.success && resp.data) {
        setUser(resp.data.user);
        setActiveOrg(resp.data.active_organisation);
        setOrganisations(resp.data.organisations || []);
      } else {
        setUser(null);
        setActiveOrg(null);
        setOrganisations([]);
      }
    } catch {
      setUser(null);
      setActiveOrg(null);
      setOrganisations([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    refreshProfile();
  }, []);

  const login = async (email: string, password: string, orgID?: string) => {
    setIsLoading(true);
    try {
      const resp = await api.auth.login({ email, password, org_id: orgID });
      if (resp.success && resp.data) {
        localStorage.setItem('afriops_token', resp.data.token);
        setUser(resp.data.user);
        setActiveOrg(resp.data.active_organisation);
        setOrganisations(resp.data.organisations || []);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (email: string, password: string, fullName: string, orgName?: string) => {
    setIsLoading(true);
    try {
      const resp = await api.auth.register({ email, password, full_name: fullName, org_name: orgName });
      if (resp.success && resp.data) {
        localStorage.setItem('afriops_token', resp.data.token);
        setUser(resp.data.user);
        setActiveOrg(resp.data.active_organisation);
        setOrganisations(resp.data.organisations || []);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      await api.auth.logout();
    } catch {
      // Ignore
    } finally {
      localStorage.removeItem('afriops_token');
      setUser(null);
      setActiveOrg(null);
      setOrganisations([]);
    }
  };

  const switchOrg = async (orgID: string) => {
    setIsLoading(true);
    try {
      const resp = await api.auth.switchOrg(orgID);
      if (resp.success && resp.data) {
        localStorage.setItem('afriops_token', resp.data.token);
        setUser(resp.data.user);
        setActiveOrg(resp.data.active_organisation);
        setOrganisations(resp.data.organisations || []);
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        activeOrg,
        organisations,
        isAuthenticated: !!user && !!activeOrg,
        isLoading,
        login,
        register,
        logout,
        switchOrg,
        refreshProfile,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
