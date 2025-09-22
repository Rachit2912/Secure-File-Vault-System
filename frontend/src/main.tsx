import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import App from "./App";
import "./styles/globals.css";

// Mount React app into #root
ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    {/* Router wraps the app for navigation */}
    <BrowserRouter>
      {/* Provide authentication state globally */}
      <AuthProvider>
        <App />
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>
);
