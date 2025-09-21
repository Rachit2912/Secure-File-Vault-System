// Full datetime (e.g., "Sep 21, 2025, 3:45 PM")
export const formatDateTime = (iso: string): string => {
  return new Date(iso).toLocaleString();
};

// Date only (e.g., "Sep 21, 2025")
export const formatDate = (iso: string): string => {
  return new Date(iso).toLocaleDateString();
};
