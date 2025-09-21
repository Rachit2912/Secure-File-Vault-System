import React from "react";

// props for storage stats :
type Props = {
  originalSize: number;
  dedupSize: number;
  saveSize: number;
};

// displays sstorage usage : original, deduped & savings :
export const StorageStats: React.FC<Props> = ({
  originalSize,
  dedupSize,
  saveSize,
}) => {
  return (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginTop: "2rem",
        gap: "2rem",
      }}
    >
      {/* Left column : Original + Deduplicated */}
      <div style={{ flex: 1, textAlign: "center" }}>
        <div style={{ marginBottom: "2rem" }}>
          <h3 style={{ margin: 0 }}>Original Size</h3>
          <p style={{ fontSize: "1.5rem", fontWeight: "bold" }}>
            {originalSize} Bytes
          </p>
        </div>
        <div>
          <h3 style={{ margin: 0 }}>Deduplicated Size</h3>
          <p style={{ fontSize: "1.5rem", fontWeight: "bold" }}>
            {dedupSize} Bytes
          </p>
        </div>
      </div>

      {/* Right column: savings */}
      <div
        style={{
          flex: 1,
          textAlign: "center",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          background: "#f0fff4",
          border: "2px solid #38a169",
          borderRadius: "12px",
          padding: "2rem",
        }}
      >
        <h2 style={{ margin: 0, fontSize: "2rem", color: "#2f855a" }}>Saved</h2>
        <p
          style={{
            margin: 0,
            fontSize: "3rem",
            fontWeight: "bold",
            color: "#2f855a",
          }}
        >
          {saveSize} Bytes
        </p>
        {/* show percentage saved relative to original :  */}
        <p style={{ margin: 0, fontSize: "1.5rem", color: "#38a169" }}>
          {originalSize ? ((saveSize / originalSize) * 100).toFixed(1) : 0}%
        </p>
      </div>
    </div>
  );
};
