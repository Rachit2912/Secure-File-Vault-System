import React from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import LoginPage from "../pages/auth/LoginPage";
import SignupPage from "../pages/auth/SignupPage";
import Dashboard from "../pages/dashboard/Dashboard";
import FileDetail from "../pages/dashboard/FileDetail";
import AdminDashboard from "../pages/admin/AdminDashboard";
import RoleManagementPage from "../pages/admin/RoleManagement";

// wrapper fn. : requires any logged-in user :
const ProtectedRoute = ({ children }: { children: JSX.Element }) => {
  const { user, loading } = useAuth();

  if (loading) return <p>Loading...</p>;
  if (!user) return <Navigate to="/login" replace />;

  return children;
};

// wrapper fn. : requires admin role :
const AdminProtectedRoute = ({ children }: { children: JSX.Element }) => {
  const { user, loading } = useAuth();

  if (loading) return <p>Loading...</p>;
  if (!user) return <Navigate to="/login" replace />;
  if (user.role !== "admin") return <Navigate to="/home" replace />;

  return children;
};

// routing :
const AppRoutes = () => {
  const { user } = useAuth();

  return (
    // all routes :

    <Routes>
      {/* auth routes :  */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/signup" element={<SignupPage />} />

      {/* root redirection : send to dashboard depending on role :  */}
      <Route
        path="/"
        element={
          user ? (
            user.role === "admin" ? (
              <Navigate to="/admin" replace />
            ) : (
              <Navigate to="/home" replace />
            )
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />

      {/* admin routes :  */}
      <Route
        path="/admin"
        element={
          <AdminProtectedRoute>
            <AdminDashboard />
          </AdminProtectedRoute>
        }
      />

      <Route
        path="/role-management"
        element={
          <AdminProtectedRoute>
            <RoleManagementPage />
          </AdminProtectedRoute>
        }
      />

      {/* user routes :  */}
      <Route
        path="/home"
        element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        }
      />

      {/* public routes :  */}
      <Route path="/fileDetails/:id" element={<FileDetail />} />

      {/* fallback : redirect unknown routes : */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

export default AppRoutes;
