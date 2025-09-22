import { useError } from "../../contexts/ErrorContext";

// global error banner shown at top of the screen :
export default function GlobalError() {
  const { error, clearError } = useError();
  if (!error) return null;

  return (
    <div
      style={{
        position: "fixed",
        top: 10,
        left: "50%",
        transform: "translateX(-50%)",
        background: "red",
        color: "white",
        fontWeight: "bold",
        padding: "10px 20px",
        borderRadius: "8px",
        zIndex: 9999,
        cursor: "pointer",
      }}
      onClick={clearError}
    >
      {error}
    </div>
  );
}
