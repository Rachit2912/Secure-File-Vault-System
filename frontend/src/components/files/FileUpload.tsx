import React, { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";
import { uploadFile } from "../../api/files";
import type { FileMeta } from "../../api/files";
import { notifyRateLimit } from "../../api/index";
import { formatBytes } from "../../utils/format";

interface Props {
  onUpaded: (file: FileMeta) => void;
  maxFileSizeBytes?: number;
}

type UploadingFile = {
  id: string;
  file: File;
  status: "pending" | "uploading" | "done" | "error";
  error?: string;
};

// file upload component with drag-and-drop and batch upload support :
export const FileUpload: React.FC<Props> = ({
  onUploaded,
  maxFileSizeBytes = 10 * 1024 * 1024,
}) => {
  const [files, setFiles] = useState<UploadingFile[]>([]);

  // 1. handle dropped files :
  // assign unique IDs, mark files too large as error else mark as pending :
  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      const next = acceptedFiles.map((f) => ({
        id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
        file: f,
        status: f.size > maxFileSizeBytes ? "error" : "pending",
        error:
          f.size > maxFileSizeBytes
            ? `Too large (max ${formatBytes(maxFileSizeBytes)})`
            : undefined,
      }));
      setFiles((prev) => [...prev, ...next]);
    },
    [maxFileSizeBytes]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    multiple: true,
  });

  // 2. upload all 'pending' marked files :
  const handleUploadAll = async () => {
    for (const f of files) {
      if (f.status !== "pending") continue;

      // a. update status to 'uploading':
      setFiles((prev) =>
        prev.map((x) => (x.id === f.id ? { ...x, status: "uploading" } : x))
      );

      try {
        // b. upload to backend :
        const meta = await uploadFile(f.file);

        // c. if success then mark as 'done' :
        setFiles((prev) =>
          prev.map((x) => (x.id === f.id ? { ...x, status: "done" } : x))
        );
        onUploaded(meta);
      } catch (err: any) {
        // d. catching errors :
        console.error("Upload failed:", err);

        let displayErr = err.message ?? "Upload failed";

        // d1. rate limit exceeded error :
        if (displayErr.toLowerCase().includes("rate limit exceeded")) {
          notifyRateLimit(displayErr);
          return;
        }

        // d2. storage quota exceeded :
        if (err.status === 403 && err.details) {
          displayErr = `${err.details.error} (Allowed: ${err.details.allowed}, Used: ${err.details.used})`;
        }

        // d3. MIME type mismatch :
        if (err.code === "StatusPreconditionFailed" || err.status === 412) {
          displayErr = "file extension does not match detected MIME type";
        }

        // d4. finalize as error :
        setFiles((prev) =>
          prev.map((x) =>
            x.id === f.id
              ? {
                  ...x,
                  status: "error",
                  error: displayErr,
                }
              : x
          )
        );
      }
    }
  };

  return (
    <div>
      {/* drag-and-drop input area :  */}
      <div
        {...getRootProps()}
        style={{
          border: "2px dashed #aaa",
          padding: 20,
          borderRadius: 8,
          textAlign: "center",
          background: isDragActive ? "#f9f9f9" : "transparent",
        }}
      >
        <input {...getInputProps()} />
        {isDragActive ? (
          <p>Drop files here...</p>
        ) : (
          <p>Drag & drop or click to select files</p>
        )}
      </div>

      {/* file list with status and errors :  */}
      <ul style={{ marginTop: 12 }}>
        {files.map((f) => (
          <li key={f.id}>
            {f.file.name} ({formatBytes(f.file.size)}) — {f.status}
            {f.error && <span style={{ color: "red" }}> – {f.error}</span>}
          </li>
        ))}
      </ul>

      {/* upload all button (visible if any file is 'pending') : */}
      {files.some((f) => f.status === "pending") && (
        <button
          onClick={handleUploadAll}
          style={{
            marginTop: 12,
            padding: "6px 12px",
            borderRadius: 6,
            border: "none",
            background: "#2b6cb0",
            color: "#fff",
            cursor: "pointer",
          }}
        >
          Upload All
        </button>
      )}
    </div>
  );
};
