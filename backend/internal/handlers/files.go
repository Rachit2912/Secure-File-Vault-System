package handlers

import (
	"backend/db"
	"backend/middleware"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// uploading-file api handler :
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Getting userId from JWT context
	uidVal := r.Context().Value(middleware.ContextUserIDKey)
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized, JWT token not found ", http.StatusUnauthorized)
		return
	}

	// Parse file form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File greater than 10 MB", http.StatusBadRequest)
		return
	}

	// get file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found in form", http.StatusNotFound)
		return
	}
	defer file.Close()

	// create unique path
	timestamp := time.Now().UnixNano()
	filePath := fmt.Sprintf("./uploads/%d_%s", timestamp, handler.Filename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Could not create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// save + hash
	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(out, hasher), file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	// insert into DB with user_id
	_, err = db.DB.Exec(
		"INSERT INTO files (user_id, filename, filepath, hash, size) VALUES ($1, $2, $3, $4, $5)",
		userID, handler.Filename, filePath, hash, size,
	)
	if err != nil {
		http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// response
	resp := map[string]string{
		"status":   "ok",
		"filename": handler.Filename,
		"hash":     hash,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// viewing-files api handler :
func FilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	//  Get userId from JWT context
	userID, ok := r.Context().Value(middleware.ContextUserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized userID", http.StatusUnauthorized)
		return
	}


	// only fetch files belonging to this user
	rows, err := db.DB.Query(
		"SELECT id, filename, size, uploaded_at FROM files WHERE user_id=$1 ORDER BY uploaded_at DESC",
		userID,
	)
	if err != nil {
		http.Error(w, "DB query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []map[string]interface{}
	for rows.Next() {
		var id int
		var filename string
		var size int64
		var uploadedAt time.Time

		err := rows.Scan(&id, &filename, &size, &uploadedAt)
		if err != nil {
			http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		files = append(files, map[string]interface{}{
			"id":          id,
			"filename":    filename,
			"size":        size,
			"uploaded_at": uploadedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
	})
}
