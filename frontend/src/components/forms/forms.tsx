import React, { useState } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { useNavigate } from "react-router-dom";

// login form component :
export const LoginForm: React.FC = () => {
  const { login } = useAuth();
  const navigate = useNavigate();

  // local states :
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // handle login submit :
  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      // 1. call login API via context :
      await login(username, password);
      // 2. redirect to root (will auto-redirect based on the role of the user) :
      navigate("/", { replace: true });
    } catch (e: any) {
      // 3. capture erroro and show messsage :
      setErr(e.message || "Login failed");
    } finally {
      // 4. always clear loading state :
      setLoading(false);
    }
  };

  return (
    <form onSubmit={onSubmit}>
      {/* error message :  */}
      {err && <div style={{ color: "red" }}>{err}</div>}

      {/* username field :  */}
      <div>
        <label>Username</label>
        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          required
        />
      </div>

      {/* password field :  */}
      <div>
        <label>Password</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
      </div>

      {/* submit button :  */}
      <button type="submit" disabled={loading}>
        {loading ? "Logging in..." : "Login"}
      </button>
    </form>
  );
};
