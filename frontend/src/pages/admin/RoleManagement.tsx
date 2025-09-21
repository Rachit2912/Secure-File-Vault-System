import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import * as adminApi from "../../api/admin";

const RoleManagementPage: React.FC = () => {
  // local state for username input, loading status & feedback message:
  const [username, setUsername] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const navigate = useNavigate();

  // step-wise role change handler :
  const handleRoleChange = async (role: "admin" | "user") => {
    // 1.validation :
    if (!username) {
      setMessage("Please enter a username");
      return;
    }
    setLoading(true);
    setMessage("");

    try {
      // 2. call backend based on selected role :
      if (role === "admin") {
        await adminApi.makeAdmin(username);
        setMessage(`${username} is now an Admin ✅`);
      } else {
        await adminApi.makeUser(username);
        setMessage(`${username} is now a User ✅`);
      }
    } catch (err: any) {
      // 3. error handling :
      setMessage("Error: " + (err.message || "Something went wrong"));
    } finally {
      // 4. always clear loading :
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-6">
      <h1 className="text-2xl font-bold mb-6">Role Management</h1>

      {/* input for username :  */}
      <input
        type="text"
        placeholder="Enter username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        className="border rounded-lg p-2 w-64 mb-4"
      />

      {/* button for making 'admin' :  */}
      <div className="flex gap-4 mb-4">
        <button
          disabled={loading}
          onClick={() => handleRoleChange("admin")}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg disabled:opacity-50"
        >
          Make Admin
        </button>

        {/* button for making 'user' :  */}
        <button
          disabled={loading}
          onClick={() => handleRoleChange("user")}
          className="bg-green-600 text-white px-4 py-2 rounded-lg disabled:opacity-50"
        >
          Make User
        </button>
      </div>

      {/* feedback message :  */}
      {message && <p className="text-sm">{message}</p>}

      {/* back navigation button :  */}
      <button
        onClick={() => navigate("/admin")}
        className="mt-6 text-gray-600 underline"
      >
        Back to Admin Dashboard
      </button>
    </div>
  );
};

export default RoleManagementPage;
