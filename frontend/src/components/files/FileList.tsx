import React from "react";
import { useNavigate } from "react-router-dom";

// displays a table of files with metadata and actions :
export const FileList: React.FC<Props> = ({ files }) => {
  const navigate = useNavigate();

  // if no files, show placeholder message :
  if (!files || files.length === 0) {
    return <p>No files uploaded yet.</p>;
  }

  return (
    <table
      style={{
        width: "100%",
        borderCollapse: "collapse",
        marginTop: "1rem",
      }}
    >
      <thead>
        <tr style={{ background: "#f4f4f4" }}>
          <th style={thStyle}>Filename</th>
          <th style={thStyle}>Uploader</th>
          <th style={thStyle}>Size</th>
          <th style={thStyle}>Uploaded At</th>
          <th style={thStyle}>Deduplicated</th>
          <th style={thStyle}>Actions</th>
        </tr>
      </thead>
      <tbody>
        {files.map((f) => (
          // each row corresponds to a specific file :
          <tr key={f.id}>
            <td style={tdStyle}>{f.filename}</td>
            <td style={tdStyle}>{f.uploader}</td>
            <td style={tdStyle}>{Math.round(f.size / 1024)} KB</td>
            <td style={tdStyle}>{new Date(f.uploaded_at).toLocaleString()}</td>
            <td style={tdStyle}>{f.deduplicated ? "Master" : "Duplicate"}</td>
            <td style={tdStyle}>
              {/* open-file button */}
              <button
                style={{
                  padding: "4px 10px",
                  border: "none",
                  borderRadius: "4px",
                  background: "#2b6cb0",
                  color: "white",
                  cursor: "pointer",
                }}
                onClick={() => navigate(`/fileDetails/${f.id}`)}
              >
                Open
              </button>

              {/* share link (only visible for public files) */}
              {f.is_public && (
                <div style={{ marginTop: "4px" }}>
                  <a
                    href={`${window.location.origin}/fileDetails/${f.id}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{ fontSize: "0.85rem", color: "#2b6cb0" }}
                  >
                    Share Link
                  </a>
                </div>
              )}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

// shared table header style :
const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "8px",
  borderBottom: "2px solid #ddd",
};

// shared table cell style :
const tdStyle: React.CSSProperties = {
  padding: "8px",
  borderBottom: "1px solid #ddd",
};
