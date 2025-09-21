package handlers

import (
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// exported constant so handlers can use it
type contextKey string

const ContextUserIDKey = contextKey("userID")

// Refresh handler - refreshes user session (get current logged-in user):
func RefershHandler(w http.ResponseWriter, r *http.Request) {

	// getting context :
	uidVal := r.Context().Value(middleware.ContextUserIDKey)
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// fetching user from DB:
	user, err := models.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// return safe user info :
	resp := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Signup handler - Registers new user :
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u models.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// hasing the pswd :
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// database entry :
	_, err = db.DB.Exec(
		"INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)",
		u.Username, u.Email, string(hashed), "user",
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

// Login handler - Authenticate user and issue JWT cookie :
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u models.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// getting hashed password from DB :
	var id int
	var hashedPwd string
	var role string
	err = db.DB.QueryRow("SELECT id, password, role FROM users WHERE username=$1", u.Username).Scan(&id, &hashedPwd, &role)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// comparing the hashes :
	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(u.Password))
	if err != nil {
		http.Error(w, "Invalid password compare hash and pswd", http.StatusUnauthorized)
		return
	}

	// generate JWT :
	token, err := services.GenerateJWT(id, u.Username, role)
	if err != nil {
		http.Error(w, "failed to generate the JWT ", http.StatusExpectationFailed)
		return
	}

	// set it to cookies :
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// sucess response :
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"msg":    "login success",
	})
}

// Logout handler - clearing JWT cookie :
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // deleting it immediately
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
