import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  getFileDetails,
  deleteFile,
  downloadFile,
  togglePrivacy,
} from "../../api/files";
import { useAuth } from "../../contexts/AuthContext";
import { formatDate, formatDateTime } from "../../utils/date";

const FileDetail: React.FC = () => {
  const { id } = useParams();
  const { user } = useAuth();
  const [file, setFile] = useState<any>(null);
  const [error, setError] = useState<string | null>(null); // 👈 error state
  const navigate = useNavigate();

  useEffect(() => {
    async function fetchFile() {
      try {
        const data = await getFileDetails(Number(id));
        setFile(data);
      } catch (err: any) {
        // Handle backend errors gracefully
        if (err.status === 403) {
          setError("🚫 This file is private. You are not allowed to view it.");
        } else if (err.status === 404) {
          setError("❌ File not found.");
        } else {
          setError("⚠️ Failed to load file details.");
        }
      }
    }
    fetchFile();
  }, [id]);

  if (error) {
    return (
      <div style={{ padding: "2rem", color: "red" }}>
        <p>{error}</p>
        <button onClick={() => navigate("/")}>Go Back</button>
      </div>
    );
  }

  if (!file) return <p>Loading...</p>;

  console.log("FileDetail response:", file);

  return (
    <div style={{ padding: "2rem" }}>
      <button onClick={() => navigate("/")}>Back</button>

      <h2>File Details</h2>
      <p>
        <strong>Filename:</strong> {file.filename}
      </p>
      <p>
        <strong>Uploader:</strong> {file.uploader_username}
      </p>
      <p>
        <strong>Size:</strong> {Math.round(file.size / 1024)} KB
      </p>
      <p>
        <strong>Uploaded At:</strong> {formatDateTime(file.uploaded_at)}
      </p>
      <p>
        <strong>Privacy:</strong> {file.is_public ? "Public" : "Private"}
      </p>
      <p>
        <strong>Download Count:</strong> {file.download_count}
      </p>

      {/* --- Public share link --- */}
      {file.is_public && (
        <p>
          <strong>Public Link:</strong>{" "}
          <a
            href={`${window.location.origin}/fileDetails/${file.id}`}
            target="_blank"
            rel="noopener noreferrer"
          >
            {`${window.location.origin}/fileDetails/${file.id}`}
          </a>
        </p>
      )}

      {/* --- Admin only - uploader details --- */}
      {user?.role?.toLowerCase() === "admin" && (
        <div style={{ marginTop: "2rem" }}>
          <h3>Uploader Details</h3>
          <p>
            <strong>Username:</strong> {file.uploader_username}
          </p>
          <p>
            <strong>Email:</strong> {file.uploader_email}
          </p>
          <p>
            <strong>Role:</strong> {file.uploader_role}
          </p>
          <p>
            <strong>Joined:</strong> {formatDate(file.uploader_created_at)}
          </p>
        </div>
      )}

      <div style={{ marginTop: "1rem" }}>
        <button onClick={() => downloadFile(file.id)}>Download</button>

        {user?.username === file.uploader_username && (
          <>
            <button
              style={{
                backgroundColor: "#c00",
                color: "white",
                marginLeft: "0.5rem",
              }}
              onClick={async () => {
                await deleteFile(file.id);
                navigate("/home");
              }}
            >
              Delete
            </button>
            <button
              style={{
                backgroundColor: "#080",
                color: "white",
                marginLeft: "0.5rem",
              }}
              onClick={async () => {
                try {
                  const res = await togglePrivacy(file.id);
                  setFile((prev: any) => ({
                    ...prev,
                    is_public: res.is_public,
                  }));
                } catch (err) {
                  console.error("Failed to toggle privacy", err);
                }
              }}
            >
              {file.is_public ? "Make Private" : "Make Public"}
            </button>
          </>
        )}
      </div>
    </div>
  );
};

export default FileDetail;
