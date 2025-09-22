package handlers

import (
	"backend/internal/db"
	"backend/internal/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// AdminFilesHandler – list all files :
func AdminFilesHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
        return
    }

    // checking for admin : 
    role, ok := r.Context().Value(middleware.ContextUserRoleKey).(string)
    if !ok || role != "admin" {
        http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
        return
    }

    // parsing filter query params :
    q := r.URL.Query()
    search := q.Get("search")
    mimeType := q.Get("mimeType")
    minSize := q.Get("minSize")
    maxSize := q.Get("maxSize")
    startDate := q.Get("startDate")
    endDate := q.Get("endDate")
    uploader := q.Get("uploader")

    // dynamic SQL query with filters :
    query := `
        SELECT f.id, f.filename, f.size, f.uploaded_at, f.is_master, f.is_public,
		u.username 
        FROM files f 
        JOIN users u ON f.user_id = u.id
		WHERE 1=1
    `
    args := []interface{}{}
    argPos := 1

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

        if err := rows.Scan(&id, &filename, &size, &uploadedAt, &isMaster,&is_public, &username); err != nil {
            http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
            return
        }

        files = append(files, map[string]interface{}{
            "id":           id,
            "filename":     filename,
            "size":         size,
            "uploaded_at":  uploadedAt.Format(time.RFC3339),
            "deduplicated": isMaster,
            "uploader":     username,
			"is_public": is_public,
        })

        // adding sizes : 
        originalSize += size
        if isMaster {
            dedupSize += size
        }
    }

    // sending  response :
    resp := map[string]interface{}{
        "files":        files,
        "dedupSize":    dedupSize,
        "originalSize": originalSize,
        "saveSize":     originalSize - dedupSize,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}


// MakeAdminHandler – change user role to admin
func MakeAdminHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
        return
    }

    // checking role from context : 
    role, ok := r.Context().Value(middleware.ContextUserRoleKey).(string)
    if !ok || role != "admin" {
        http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
        return
    }

    // parsing request body for filters  :
    var req struct {
        Username string `json:"username"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // updating entry in DB :
    _, err := db.DB.Exec("UPDATE users SET role = $1 WHERE username = $2", "admin", req.Username)
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // response : 
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "ok",
        "username": req.Username,
        "newRole":  "admin",
    })
}

// MakeUserHandler – change user role to normal user
func MakeUserHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
        return
    }

    // checking for admin role : 
    role, ok := r.Context().Value(middleware.ContextUserRoleKey).(string)
    if !ok || role != "admin" {
        http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
        return
    }

    // parsing request body : 
    var req struct {
        Username string `json:"username"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // updating entry in DB : 
    _, err := db.DB.Exec("UPDATE users SET role = $1 WHERE username = $2", "user", req.Username)
    if err != nil {
        http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // sending response : 
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "ok",
        "username": req.Username,
        "newRole":  "user",
    })
}
