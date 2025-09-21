import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import * as authApi from "../api/auth";

// sturcture for authenticated user :
export type User = {
  id: number;
  username: string;
  email?: string;
  role: string;
};

// contract for what AuthContext provides :
type AuthContextType = {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  signup: (username: string, email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
};

// creating a context :
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// hook to consume context :
export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
};

// provider for authentication state & actions:
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // refresh current user from backend :
  const refresh = async () => {
    try {
      // 1. call /api/me :
      const data = await authApi.apiMe();
      // 2. normalize response :
      const u = data && (data.user ?? data);
      // 3. update state :
      setUser({
        id: u.id,
        username: u.username,
        email: u.email,
        role: u.role,
      });
    } catch (err: any) {
      // unauthorized on error then clear user :
      if (err?.status === 401) {
        setUser(null);
      } else {
        setUser(null);
      }
    } finally {
      setLoading(false);
    }
  };

  // on mounting, trying restoring session :
  useEffect(() => {
    refresh();
  }, []);

  // login flow :
  const login = async (username: string, password: string) => {
    setLoading(true);
    try {
      const res = await authApi.apiLogin({ username, password });
      await refresh();
    } finally {
      setLoading(false);
    }
  };

  // Signup flow :
  const signup = async (username: string, email: string, password: string) => {
    setLoading(true);
    try {
      const res = await authApi.apiSignup({ username, email, password });
      await refresh();
    } finally {
      setLoading(false);
    }
  };

  // logout flow :
  const logout = async () => {
    setLoading(true);
    try {
      await authApi.apiLogout();
      setUser(null);
    } finally {
      setUser(null);
      setLoading(false);
    }
  };

  // context value :
  const value = { user, loading, login, signup, logout, refresh };
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
