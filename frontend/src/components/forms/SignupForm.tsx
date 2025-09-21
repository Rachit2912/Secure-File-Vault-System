import React, { useState } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { useNavigate } from "react-router-dom";

// signup form component :
export const SignupForm: React.FC = () => {
  const { signup } = useAuth();
  const navigate = useNavigate();

  // local states :
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // handle signup submit :
  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      // 1. call signup API via context :
      await signup(username, email, password);
      // 2. redirect to root (will auto-redirect based on the role of the user) :
      navigate("/home");
    } catch (e: any) {
      // 3. capture erroro and show messsage :
      setErr(e.message || "Signup failed");
    } finally {
      // 4. always clear loading state :
      setLoading(false);
    }
  };

  return (
    <form onSubmit={onSubmit}>
      {/* error message */}
      {err && <div style={{ color: "red" }}>{err}</div>}

      {/* username field : */}
      <div>
        <label>Username</label>
        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          required
        />
      </div>

      {/* email field : */}
      <div>
        <label>Email</label>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
      </div>

      {/* password field : */}
      <div>
        <label>Password</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
      </div>

      {/* submit button : */}
      <button type="submit" disabled={loading}>
        {loading ? "Signing up..." : "Signup"}
      </button>
    </form>
  );
};
