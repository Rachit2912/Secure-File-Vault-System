// formatBytes: convert raw bytes into human-readable units
export const formatBytes = (bytes: number) => {
  // special case: zero bytes
  if (bytes === 0) return "0 B";

  // define conversion base (1 KB = 1024 B)
  const k = 1024;

  // supported units
  const sizes = ["B", "KB", "MB", "GB"];

  // 1: find which unit to use (logarithmic)
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  // step 2: scale the number and format with 2 decimals
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};
