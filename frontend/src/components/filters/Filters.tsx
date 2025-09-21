import React, { useState } from "react";

// structure of filter values :
export interface FilterValues {
  search: string;
  mimeType: string;
  minSize: string;
  maxSize: string;
  startDate: string;
  endDate: string;
  uploader: string;
}

// props with 'apply' and 'reset' button for filters :
interface FiltersProps {
  onApply: (filters: FilterValues) => void;
  onReset?: () => void;
}

// filters component for searching/filtering files :
const Filters: React.FC<FiltersProps> = ({ onApply, onReset }) => {
  // local state for filters :
  const [filters, setFilters] = useState<FilterValues>({
    search: "",
    mimeType: "",
    minSize: "",
    maxSize: "",
    startDate: "",
    endDate: "",
    uploader: "",
  });

  // handle input changes :
  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFilters((prev) => ({ ...prev, [name]: value }));
  };

  // handle submit button  functionality :
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onApply(filters);
  };

  // handle reset functionality :
  const handleReset = () => {
    setFilters({
      search: "",
      mimeType: "",
      minSize: "",
      maxSize: "",
      startDate: "",
      endDate: "",
      uploader: "",
    });
    onReset?.();
  };

  return (
    <form
      onSubmit={handleSubmit}
      style={{
        display: "flex",
        alignItems: "center",
        gap: "8px",
        whiteSpace: "nowrap",
        overflowX: "auto",
        padding: "8px",
        border: "1px solid #ccc",
        borderRadius: "6px",
        background: "#f9f9f9",
      }}
    >
      {/* search by filename :  */}
      <input
        type="text"
        name="search"
        placeholder="Filename"
        value={filters.search}
        onChange={handleChange}
        style={{ padding: "4px", width: "100px" }}
      />

      {/* MIME type filter :  */}
      <select
        name="mimeType"
        value={filters.mimeType}
        onChange={handleChange}
        style={{ padding: "4px", width: "90px" }}
      >
        <option value="">Type</option>
        <option value="image/jpeg">JPEG</option>
        <option value="image/png">PNG</option>
        <option value="application/pdf">PDF</option>
        <option value="text/plain">Text</option>
        <option value="application/zip">ZIP</option>
      </select>

      {/* min file size range :  */}
      <input
        type="number"
        name="minSize"
        placeholder="Min KB"
        value={filters.minSize}
        onChange={handleChange}
        style={{ padding: "4px", width: "70px" }}
      />
      {/* max file size range :  */}
      <input
        type="number"
        name="maxSize"
        placeholder="Max KB"
        value={filters.maxSize}
        onChange={handleChange}
        style={{ padding: "4px", width: "70px" }}
      />

      {/* starting date range: */}
      <input
        type="date"
        name="startDate"
        value={filters.startDate}
        onChange={handleChange}
        style={{ padding: "4px", width: "130px" }}
      />
      {/* ending date range: */}
      <input
        type="date"
        name="endDate"
        value={filters.endDate}
        onChange={handleChange}
        style={{ padding: "4px", width: "130px" }}
      />

      {/* uploader filter: */}
      <input
        type="text"
        name="uploader"
        placeholder="Uploader"
        value={filters.uploader}
        onChange={handleChange}
        style={{ padding: "4px", width: "90px" }}
      />

      {/* apply button: */}
      <button type="submit" style={{ padding: "4px 8px" }}>
        Apply
      </button>

      {/* reset button: */}
      <button
        type="button"
        onClick={handleReset}
        style={{ padding: "4px 8px" }}
      >
        Reset
      </button>
    </form>
  );
};

export default Filters;
