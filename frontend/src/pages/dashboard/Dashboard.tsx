import React, { useEffect, useState } from "react";
import { useAuth } from "../../contexts/AuthContext";
import { listFiles } from "../../api/files";
import type { FileMeta } from "../../api/files";
import { FileUpload } from "../../components/files/FileUpload";
import { FileList } from "../../components/files/FileList";
import { StorageStats } from "../../components/stats/StorageStats";
import { listPublicFiles } from "../../api/files";
import { PublicFileList } from "../../components/files/PublicFileList";
import type { PublicFile } from "../../api/files";
import Filters from "../../components/filters/Filters";

const Dashboard: React.FC = () => {
  const { user, logout } = useAuth();

  // state for pvt files :
  const [files, setFiles] = useState<FileMeta[]>([]);
  const [loading, setLoading] = useState(true);
  const [originalSize, setOriginalSize] = useState(0);
  const [dedupSize, setDedupSize] = useState(0);
  const [saveSize, setSaveSize] = useState(0);

  // state for public files :
  const [publicFiles, setPublicFiles] = useState<PublicFile[]>([]);
  const [publicTotal, setPublicTotal] = useState(0);

  // fetch both pvt and public files:
  async function fetchFiles() {
    try {
      // getting pvt files :
      const res = await listFiles();
      setFiles(res.files ?? []);
      setOriginalSize(res.originalSize ?? 0);
      setDedupSize(res.dedupSize ?? 0);
      setSaveSize(res.saveSize ?? 0);

      // getting all public files :
      let pubRes = await listPublicFiles();
      setPublicFiles(pubRes.files ?? []);
      setPublicTotal(pubRes.total ?? 0);
    } catch (err) {
      console.error("Failed to fetch files:", err);
    } finally {
      setLoading(false);
    }
  }

  // on mounting, run it immediatelyy:
  useEffect(() => {
    fetchFiles();
  }, []);
  if (!user) return <p>Not authorized</p>;

  return (
    <div style={{ padding: "2rem" }}>
      <h1>Welcome, {user.username} 👋</h1>
      <button onClick={logout}>Logout</button>

      <hr />

      {/* file upload section :  */}
      <h2>Upload a new file</h2>
      <FileUpload onUploaded={() => fetchFiles()} />

      {/* user - pvt files section:  */}
      <h2>Your Files</h2>
      {loading ? (
        <p>Loading files...</p>
      ) : (
        <div>
          {/* filter bar :  */}
          <Filters
            onApply={(filters) => {
              console.log("Applied filters:", filters);
              listFiles(filters).then((res) => {
                setFiles(res.files ?? []);
                setOriginalSize(res.originalSize ?? 0);
                setDedupSize(res.dedupSize ?? 0);
                setSaveSize(res.saveSize ?? 0);
              });
            }}
            onReset={() => {
              fetchFiles();
            }}
          />

          {/* list user files :  */}
          <FileList
            files={files}
            onDeleted={(id) =>
              setFiles((prev) => prev.filter((file) => file.id !== id))
            }
          />
        </div>
      )}

      <hr />

      {/* public files section :  */}
      <h2>Public Files ({publicTotal})</h2>
      {loading ? (
        <p>Loading public files...</p>
      ) : (
        <PublicFileList files={publicFiles} />
      )}

      <hr />

      {/* storage stasticis :  */}
      <StorageStats
        originalSize={originalSize}
        dedupSize={dedupSize}
        saveSize={saveSize}
      />
    </div>
  );
};

export default Dashboard;
