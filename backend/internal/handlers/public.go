package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/db"
)

// structure for public listing files :
type PublicFile struct {
	ID            int       `json:"id"`
	Filename      string    `json:"filename"`
	Size          int64     `json:"size"`
	UploadedAt    time.Time `json:"uploaded_at"`
	IsMaster      bool      `json:"is_master"`
	Uploader      string    `json:"uploader"`
	DownloadCount int       `json:"download_count"`
}

// PublicFilesHandler - gets all public files :
func PublicFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// executing the query :
	rows, err := db.DB.Query(`
		SELECT f.id, f.filename, f.size, f.uploaded_at, f.is_master, u.username, f.download_count
		FROM files f
		JOIN users u ON f.user_id = u.id
		WHERE f.is_public = TRUE
		ORDER BY f.uploaded_at DESC
	`)
	if err != nil {
		http.Error(w, "DB query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// preparing the response :
	files := make([]PublicFile, 0)
	for rows.Next() {
		var pf PublicFile
		if err := rows.Scan(&pf.ID, &pf.Filename, &pf.Size, &pf.UploadedAt, &pf.IsMaster, &pf.Uploader, &pf.DownloadCount); err != nil {
			http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, pf)
	}

	// adding total count :
	var total int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM files WHERE is_public = TRUE`).Scan(&total); err != nil && err != sql.ErrNoRows {
		http.Error(w, "DB count error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// sending response : 
	resp := map[string]interface{}{
		"files": files,
		"total": total,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
