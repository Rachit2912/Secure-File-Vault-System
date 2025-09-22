import React, { createContext, useContext, useState, useEffect } from "react";

// structure for error context :
type ErrorContextType = {
  error: string | null;
  clearError: () => void;
};

// creating context :
const ErrorContext = createContext<ErrorContextType | undefined>(undefined);

// provider to handle global error state :
export const ErrorProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [error, setError] = useState<string | null>(null);

  const clearError = () => setError(null);

  // 1. listen for global 'rateLimitError' events :
  useEffect(() => {
    const handler = (e: any) => setError(e.detail);
    window.addEventListener("rateLimitError", handler);
    return () => window.removeEventListener("rateLimitError", handler);
  }, []);

  // 2. auto-hide after 2 seconds :
  useEffect(() => {
    if (error) {
      const t = setTimeout(() => setError(null), 2000);
      return () => clearTimeout(t);
    }
  }, [error]);

  return (
    <ErrorContext.Provider value={{ error, clearError }}>
      {children}
    </ErrorContext.Provider>
  );
};

// hook to consume error context :
export const useError = () => {
  const ctx = useContext(ErrorContext);
  if (!ctx) throw new Error("useError must be used inside ErrorProvider");
  return ctx;
};
