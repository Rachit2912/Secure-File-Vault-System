package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// User data-structure :
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var db *sql.DB


func main() {
	// connection to postgreSQL database : 
	connStr := "postgres://filevault_db:filevault_db@localhost:5432/filevault?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to DB:", err)
	}
	fmt.Println("✅ Connected to Postgres")

	// Ensure uploads dir exists
	os.MkdirAll("./uploads", os.ModePerm)

	// Routing done : 
	mux := http.NewServeMux()
	mux.HandleFunc("/api/me", demoCheck)
	mux.HandleFunc("/api/signup", signupHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/upload", uploadHandler)
	mux.HandleFunc("/api/files", filesHandler)



	fmt.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(mux)))
}

// demo api checkup : 
func demoCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// signup api handler : 
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// password-hashing : 
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// database entries : 
	_, err = db.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		u.Username, u.Email, string(hashed),
	)
	if err != nil {
		http.Error(w, "User already exists or DB error: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"msg":    "user created",
	})
}

// login api handler : 
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// fmt.Println("DEBUG - Incoming login payload:", u.Username, u.Password)


	// getting stored password hash
	var hashedPwd string
	err = db.QueryRow("SELECT password FROM users WHERE username=$1", u.Username).Scan(&hashedPwd)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// checking both pswds 
	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(u.Password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// login sucess : 
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"msg":    "login success",
	})
}

// upload-file api handler : 
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse file form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// getting the file 
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found in form", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Create unique paths for new files : 
	timestamp := time.Now().UnixNano()
	filePath := fmt.Sprintf("./uploads/%d_%s", timestamp, handler.Filename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Could not create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Saving + hashing 
	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(out, hasher), file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusExpectationFailed)
		return
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	// database entry : 
	_, err = db.Exec(
		"INSERT INTO files (filename, filepath, hash, size) VALUES ($1, $2, $3, $4)",
		handler.Filename, filePath, hash, size,
	)
	if err != nil {
		http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// response to frontend : 
	resp := map[string]string{
		"status":   "ok",
		"filename": handler.Filename,
		"hash":     hash,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// corsMiddleware for wrapping all api routes for CORS from frontend : 
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials","true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// viewing-file api handler : 
func filesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// query the database : 
	rows, err := db.Query("SELECT id, filename, size, uploaded_at FROM files ORDER BY uploaded_at DESC")
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