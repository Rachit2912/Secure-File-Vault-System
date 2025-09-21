import { getReq, handleUpload } from "./index";
import type { FilterValues } from "../components/filters/Filters";

// File Metadata type for typed responses :
export type FileMeta = {
  id: number;
  filename: string;
  size: number;
  uploaded_at: string;
  uploader: string;
  deduplicated: boolean;
  is_public: boolean;
};

// fetch files (with optional filters):
export async function listFiles(filters?: Partial<FilterValues>) {
  let url = "/api/files";

  // 1. if filters exist, convert to query params :
  if (filters) {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value) params.append(key, value.toString());
    });
    url += "?" + params.toString();
  }

  // 2. call backend and return JSON :
  return getReq(url);
}

// upload file function :
export async function uploadFile(file: File): Promise<FileMeta> {
  // 1. prepare multipart/form-data :
  const formData = new FormData();
  formData.append("file", file);

  // 2. send to backend with using helper function :
  const data = await handleUpload("/api/upload", formData);

  // 3. return typed response :
  return data as FileMeta;
}

// fetch file details by ID :
export async function getFileDetails(fileId: number) {
  return getReq(`/api/fileDetails/${fileId}`);
}

// trigger file download (redirects browser):
export async function downloadFile(fileId: number) {
  window.location.href = `/api/fileDownload/${fileId}`;
}

// toggle file public/private state :
export async function togglePrivacy(fileId: number) {
  return getReq(`/api/fileTogglePrivacy/${fileId}`);
}

// delete a file by ID :
export async function deleteFile(fileId: number) {
  return getReq(`/api/fileDelete/${fileId}`);
}

// public file listing type :
export type PublicFile = {
  id: number;
  filename: string;
  size: number;
  uploaded_at: string;
  is_master: boolean;
  uploader: string;
  download_count: number;
};

// fetch all public files :
export async function listPublicFiles(): Promise<{
  files: PublicFile[];
  total: number;
}> {
  return getReq(`/api/publicFiles`);
}
