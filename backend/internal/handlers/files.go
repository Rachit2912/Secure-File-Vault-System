package handlers

import (
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// file handler - gets user's own files :
func FilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	//  getting userID from JWT context :
	userID, ok := r.Context().Value(middleware.ContextUserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized userID", http.StatusUnauthorized)
		return
	}

	//  parsing query params for filters :
	q := r.URL.Query()
	search := q.Get("search")
	mimeType := q.Get("mimeType")
	minSize := q.Get("minSize")
	maxSize := q.Get("maxSize")
	startDate := q.Get("startDate")
	endDate := q.Get("endDate")
	uploader := q.Get("uploader")

	// writing dynamic SQL query with filters :
	query := `
		SELECT f.id, f.filename, f.size, f.uploaded_at, f.is_master, f.is_public,
		u.username 
		FROM files f 
		JOIN users u ON f.user_id = u.id
		WHERE f.user_id = $1
	`
	args := []interface{}{userID}
	argPos := 2

	if search != "" {
		query += fmt.Sprintf(" AND f.filename ILIKE $%d", argPos)
		args = append(args, "%"+search+"%")
		argPos++
	}
	if mimeType != "" {
		query += fmt.Sprintf(" AND f.mime_type = $%d", argPos)
		args = append(args, mimeType)
		argPos++
	}
	if minSize != "" {
		if kb, err := strconv.ParseInt(minSize, 10, 64); err == nil {
			bytes := kb * 1024
			query += fmt.Sprintf(" AND f.size >= $%d", argPos)
			args = append(args, bytes)
			argPos++
		}
	}
	if maxSize != "" {
		if kb, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			bytes := kb * 1024
			query += fmt.Sprintf(" AND f.size <= $%d", argPos)
			args = append(args, bytes)
			argPos++
		}
	}
	if startDate != "" {
		query += fmt.Sprintf(" AND f.uploaded_at >= $%d", argPos)
		args = append(args, startDate)
		argPos++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND f.uploaded_at <= $%d", argPos)
		args = append(args, endDate)
		argPos++
	}
	if uploader != "" {
		query += fmt.Sprintf(" AND u.username ILIKE $%d", argPos)
		args = append(args, "%"+uploader+"%")
		argPos++
	}

	query += " ORDER BY f.uploaded_at DESC"

	// executing query :
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "DB query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// processing results :
	var files []map[string]interface{}
	var originalSize, dedupSize int64
	for rows.Next() {
		var id int
		var filename string
		var size int64
		var uploadedAt time.Time
		var isMaster bool
		var username string
		var is_public bool

		if err := rows.Scan(&id, &filename, &size, &uploadedAt, &isMaster, &is_public, &username); err != nil {
			http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// preparing response :
		files = append(files, map[string]interface{}{
			"id":           id,
			"filename":     filename,
			"size":         size,
			"uploaded_at":  uploadedAt.Format(time.RFC3339),
			"deduplicated": isMaster,
			"uploader":     username,
			"is_public":    is_public,
		})

		// with adding sizes :
		originalSize += size
		if isMaster {
			dedupSize += size
		}
	}

	// sending messsages :
	resp := map[string]interface{}{
		"files":        files,
		"dedupSize":    dedupSize,
		"originalSize": originalSize,
		"saveSize":     originalSize - dedupSize,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// uploader handler - uploads the files in db :
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// getting userDI from context :
	uidVal := r.Context().Value(middleware.ContextUserIDKey)
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized, JWT token not found", http.StatusUnauthorized)
		return
	}

	// parsing form & files :
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit for files
	if err != nil {
		http.Error(w, "File greater than 10 MB", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found in form", http.StatusNotFound)
		return
	}
	defer file.Close()

	// doing mime validation :
	buf := make([]byte, 512)
	_, _ = file.Read(buf)
	if err := utils.ValidateMIME(handler.Filename, buf); err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	_, _ = file.Seek(0, io.SeekStart)

	// detecting MIME + size :
	mimeType := http.DetectContentType(buf)
	stat, _ := handler.Open()
	size, _ := stat.Seek(0, io.SeekEnd)
	_, _ = stat.Seek(0, io.SeekStart)

	// querying DB for possible duplicates (same MIME + size, master only) :
	rows, err := db.DB.Query(
		`SELECT id, user_id, filename, filepath, hash, size, mime_type, reference_count, is_master
		 FROM files
		 WHERE mime_type=$1 AND size=$2 AND is_master=TRUE`,
		mimeType, size,
	)
	if err != nil {
		http.Error(w, "DB query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var candidates []*models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.UserID, &f.Filename, &f.Filepath, &f.Hash,
			&f.Size, &f.MimeType, &f.ReferenceCount, &f.IsMaster); err != nil {
			http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		candidates = append(candidates, &f)
	}

	// Hash + compare against candidates :
	dup, hash, err := services.FindDuplicate(file, candidates)
	if err != nil {
		http.Error(w, "Error checking duplicates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = file.Seek(0, io.SeekStart) // rewind for saving if new

	if dup != nil {
		// Duplicate found: insert metadata + add ref count
		_, err = db.DB.Exec(
			`INSERT INTO files (user_id, filename, filepath, hash, size, mime_type, is_master)
			 VALUES ($1, $2, $3, $4, $5, $6, FALSE)`,
			userID, handler.Filename, dup.Filepath, hash, size, mimeType,
		)
		if err != nil {
			http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = db.DB.Exec(`UPDATE files SET reference_count = reference_count + 1 WHERE id=$1`, dup.ID)

		resp := map[string]string{"status": "duplicate-linked", "hash": hash}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// quota checking  :
	var used int64
	err = db.DB.QueryRow(`SELECT COALESCE(SUM(size),0) FROM files WHERE user_id=$1`, userID).Scan(&used)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// quota resulting :
	quota := utils.GetUserQuotaBytes()
	if used+size > quota {
		resp := map[string]interface{}{
			"error":   "Storage quota exceeded",
			"allowed": fmt.Sprintf("%d MB", quota/1024/1024),
			"used":    fmt.Sprintf("%.2f MB", float64(used+size)/1024.0/1024.0),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// New file: saved physically :
	timestamp := time.Now().UnixNano()
	filePath := fmt.Sprintf("./uploads/%d_%s", timestamp, handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Could not create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Inserted as master file  :
	_, err = db.DB.Exec(
		`INSERT INTO files (user_id, filename, filepath, hash, size, mime_type, reference_count, is_master)
		 VALUES ($1, $2, $3, $4, $5, $6, 1, TRUE)`,
		userID, handler.Filename, filePath, hash, size, mimeType,
	)
	if err != nil {
		http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"status": "new-upload", "hash": hash}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// delete handler - deletes a specific file :
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// Auth: only owner can delete :
	uidVal := r.Context().Value(middleware.ContextUserIDKey)
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract file ID :
	vars := mux.Vars(r)
	id := vars["id"]

	// Lookup file :
	var uploaderID int
	var filepathOnDisk string
	var isMaster bool
	var refCount int
	err := db.DB.QueryRow(
		`SELECT user_id, filepath, is_master, reference_count 
		 FROM files WHERE id=$1`, id,
	).Scan(&uploaderID, &filepathOnDisk, &isMaster, &refCount)

	if err == sql.ErrNoRows {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ensuring only uploader can delete :
	if uploaderID != userID {
		http.Error(w, "Forbidden: not file owner", http.StatusForbidden)
		return
	}

	// Case 1: Not master → just remove row + decrement master's ref_count :
	if !isMaster {
		// Decrement reference count of master
		_, _ = db.DB.Exec(`
			UPDATE files 
			SET reference_count = reference_count - 1 
			WHERE hash = (SELECT hash FROM files WHERE id=$1) 
			  AND is_master = TRUE`, id)

		_, err = db.DB.Exec(`DELETE FROM files WHERE id=$1`, id)
		if err != nil {
			http.Error(w, "DB delete error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Case 2: Master file :
		if refCount > 1 {
			// Pick a duplicate to promote as new master :
			var newMasterID int
			err := db.DB.QueryRow(`
				SELECT id FROM files 
				WHERE hash = (SELECT hash FROM files WHERE id=$1) 
				  AND is_master = FALSE 
				LIMIT 1`, id).Scan(&newMasterID)
			if err != nil {
				http.Error(w, "Could not find duplicate to promote: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Promote duplicate → master :
			_, _ = db.DB.Exec(`
				UPDATE files SET is_master = TRUE, filepath = $1 
				WHERE id=$2`, filepathOnDisk, newMasterID)

			// Delete current master row :
			_, _ = db.DB.Exec(`DELETE FROM files WHERE id=$1`, id)
		} else {
			// Case 3: Only reference → delete DB row + physical file :
			_, err = db.DB.Exec(`DELETE FROM files WHERE id=$1`, id)
			if err != nil {
				http.Error(w, "DB delete error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = os.Remove(filepathOnDisk) // remove file physically
		}
	}
	// Responding success :
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// fileDownloadHandler - downloades the files :
func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing file id", http.StatusBadRequest)
		return
	}

	// looking up for the file in DB :
	var filename, filepathOnDisk string
	err := db.DB.QueryRow(
		`SELECT filename, filepath FROM files WHERE id=$1`, id,
	).Scan(&filename, &filepathOnDisk)

	if err == sql.ErrNoRows {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// checking if file exists on the filePath :
	if _, err := os.Stat(filepathOnDisk); os.IsNotExist(err) {
		http.Error(w, "File missing on server", http.StatusInternalServerError)
		return
	}

	// increasing the download_count :
	_, _ = db.DB.Exec(`UPDATE files SET download_count = download_count + 1 WHERE id=$1`, id)

	// sending the response :
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filepath.Clean(filepathOnDisk))
}

// privacy change handler - changes a file's privacy  :
func FileTogglePrivacyHandler(w http.ResponseWriter, r *http.Request) {

	// checking JWT tokens :
	uidVal := r.Context().Value(middleware.ContextUserIDKey)
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract file ID
	vars := mux.Vars(r)
	id := vars["id"]

	// checking the file in DB :
	var uploaderID int
	var isPublic bool
	err := db.DB.QueryRow(
		`SELECT user_id, is_public FROM files WHERE id=$1`, id,
	).Scan(&uploaderID, &isPublic)

	if err == sql.ErrNoRows {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// making sure only uploader can toggle :
	if uploaderID != userID {
		http.Error(w, "Not allowed", http.StatusForbidden)
		return
	}

	// fkipping the privacy :
	newPrivacy := !isPublic
	_, err = db.DB.Exec(`UPDATE files SET is_public=$1 WHERE id=$2`, newPrivacy, id)
	if err != nil {
		http.Error(w, "Failed to update privacy", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"success":   true,
		"is_public": newPrivacy,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// file details handler - gives details about a particular file :
func FileDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// extracting ID from URL path :
	vars := mux.Vars(r)
	idStr := vars["id"]

	fileID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// getting file by its id :
	file, err := services.GetFileByID(fileID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// visibility check :
	// Try to read user context :
	uidVal := r.Context().Value(middleware.ContextUserIDKey)
	role, _ := r.Context().Value(middleware.ContextUserRoleKey).(string)

	var userID int
	if uidVal != nil {
		userID, _ = uidVal.(int)
	}

	// If file is NOT public, allow only uploader or admin :
	if !file.IsPublic {
		if role != "admin" && file.UploaderID != userID {
			http.Error(w, "Forbidden: private file", http.StatusForbidden)
			return
		}
	}

	// Respond with JSON :
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}
