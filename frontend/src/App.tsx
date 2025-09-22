import React from "react";
import AppRoutes from "./routes/AppRoutes";
import { ErrorProvider } from "./contexts/ErrorContext";
import GlobalError from "./components/errors/GlobalError";

// Root of the application
function App() {
  return (
    // Wrap everything in ErrorProvider so global error state is accessible
    <ErrorProvider>
      {/* Display global errors (if any) */}
      <GlobalError />

      {/* Main application routes */}
      <AppRoutes />
    </ErrorProvider>
  );
}

export default App;
