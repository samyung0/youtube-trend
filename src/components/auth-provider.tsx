"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  type ReactNode,
} from "react";
import { getMe, setToken, clearToken, getGoogleLoginURL, type User } from "@/lib/api";

interface AuthState {
  user: User | null;
  isPro: boolean;
  loading: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthState>({
  user: null,
  isPro: false,
  loading: true,
  login: () => {},
  logout: () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isPro, setIsPro] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchUser = useCallback(async () => {
    try {
      const data = await getMe();
      setUser(data.user);
      setIsPro(data.is_pro);
    } catch {
      setUser(null);
      setIsPro(false);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (token) {
      fetchUser();
    } else {
      setLoading(false);
    }

    const handleUnauth = () => {
      setUser(null);
      setIsPro(false);
    };
    window.addEventListener("auth:unauthorized", handleUnauth);
    return () => window.removeEventListener("auth:unauthorized", handleUnauth);
  }, [fetchUser]);

  const login = () => {
    window.location.href = getGoogleLoginURL();
  };

  const logout = () => {
    clearToken();
    setUser(null);
    setIsPro(false);
  };

  return (
    <AuthContext.Provider value={{ user, isPro, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}

export function ProGate({ children, fallback }: { children: ReactNode; fallback?: ReactNode }) {
  const { isPro, loading, login } = useAuth();
  if (loading) return null;
  if (!isPro) {
    return (
      fallback ?? (
        <div className="flex flex-col items-center justify-center gap-4 py-20">
          <p className="text-lg font-medium">Pro feature</p>
          <p className="text-sm text-muted-foreground">
            Upgrade to Pro to unlock this feature.
          </p>
          <button
            onClick={login}
            className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
          >
            Sign in with Google
          </button>
        </div>
      )
    );
  }
  return <>{children}</>;
}
