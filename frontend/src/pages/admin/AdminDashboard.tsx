import React, { useEffect, useState } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { FileUpload } from "../../components/files/FileUpload";
import { FileList } from "../../components/files/FileList";
import { StorageStats } from "../../components/stats/StorageStats";
import Filters from "../../components/filters/Filters";
import * as adminApi from "../../api/admin"; // new file for admin APIs
import type { FileMeta } from "../../api/files";
import { useNavigate } from "react-router-dom";

// admin dashboard component :
const AdminDashboard: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  // state for files + storage stats :
  const [files, setFiles] = useState<FileMeta[]>([]);
  const [loading, setLoading] = useState(true);
  const [originalSize, setOriginalSize] = useState(0);
  const [dedupSize, setDedupSize] = useState(0);
  const [saveSize, setSaveSize] = useState(0);

  // 1. fetch files from backend (applied filters also) :
  async function fetchFiles(filters?: any) {
    try {
      const res = await adminApi.listAdminFiles(filters);
      // 2. update files + stats :
      setFiles(res.files ?? []);
      setOriginalSize(res.originalSize ?? 0);
      setDedupSize(res.dedupSize ?? 0);
      setSaveSize(res.saveSize ?? 0);
    } catch (err) {
      console.error("Failed to fetch admin files:", err);
    } finally {
      setLoading(false);
    }
  }

  // on mounting, run it once :
  useEffect(() => {
    fetchFiles();
  }, []);

  // not found then raise error :
  if (!user) return <p>Not authorized</p>;

  return (
    <div style={{ padding: "2rem" }}>
      {/* greeting msg : */}
      <h1>Welcome Admin, {user.username} 👑</h1>
      {/* logout button : */}
      <button onClick={logout}>Logout</button>
      {/* role management button : */}
      <button onClick={() => navigate("/role-management")}>Manage Roles</button>
      <hr />

      {/* file upload section :  */}
      <h2>Upload a new file</h2>
      <FileUpload onUploaded={() => fetchFiles()} />
      {/* list of all files :  */}
      <h2>All Files</h2>
      {loading ? (
        <p>Loading files...</p>
      ) : (
        <div>
          {/* applying filters :  */}
          <Filters
            onApply={(filters) => {
              fetchFiles(filters);
            }}
            onReset={() => {
              fetchFiles();
            }}
          />

          <FileList
            files={files}
            // remove deleted file from UI without reloading all :
            onDeleted={(id) =>
              setFiles((prev) => prev.filter((file) => file.id !== id))
            }
          />
        </div>
      )}
      <hr />

      {/* storage stats preview :  */}
      <StorageStats
        originalSize={originalSize}
        dedupSize={dedupSize}
        saveSize={saveSize}
      />
    </div>
  );
};

export default AdminDashboard;
