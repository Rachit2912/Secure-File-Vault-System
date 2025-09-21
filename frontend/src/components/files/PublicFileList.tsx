import React from "react";
import { useNavigate } from "react-router-dom";
import type { PublicFile } from "../../api/files";

type Props = {
  files: PublicFile[];
};

// table to display list of all public files :
export const PublicFileList: React.FC<Props> = ({ files }) => {
  const navigate = useNavigate();

  // empty state (no public files) :
  if (!files || files.length === 0) {
    return <p>No public files available.</p>;
  }

  return (
    <table
      style={{ width: "100%", borderCollapse: "collapse", marginTop: "1rem" }}
    >
      <thead>
        <tr>
          {/* column headers :  */}
          <th style={thStyle}>Filename</th>
          <th style={thStyle}>Uploader</th>
          <th style={thStyle}>Size</th>
          <th style={thStyle}>Uploaded At</th>
          <th style={thStyle}>Downloads</th>
          <th style={thStyle}>Actions</th>
        </tr>
      </thead>
      <tbody>
        {files.map((f) => (
          <tr key={f.id}>
            {/* file information :  */}
            <td style={tdStyle}>{f.filename}</td>
            <td style={tdStyle}>{f.uploader}</td>
            <td style={tdStyle}>{Math.round(f.size / 1024)} KB</td>
            <td style={tdStyle}>{new Date(f.uploaded_at).toLocaleString()}</td>
            <td style={tdStyle}>{f.download_count}</td>

            {/* actions : navigate to file details :  */}
            <td style={tdStyle}>
              <button
                style={buttonStyle}
                onClick={() => navigate(`/fileDetails/${f.id}`)}
              >
                Open
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

// reusable css styles :
const thStyle: React.CSSProperties = {
  borderBottom: "1px solid #ccc",
  textAlign: "left",
  padding: "8px",
};

const tdStyle: React.CSSProperties = {
  borderBottom: "1px solid #eee",
  padding: "8px",
};

const buttonStyle: React.CSSProperties = {
  background: "#2b6cb0",
  color: "white",
  border: "none",
  padding: "4px 10px",
  borderRadius: "4px",
};
