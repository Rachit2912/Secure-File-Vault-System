package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gorilla/mux"
)

func main() {

	// loading environment variables from .env file
	config.LoadConfig()
	
	// connection to postgrSQL : 
	if err := db.Connect(); err != nil {
		log.Fatal("DB connection failed:", err)
	}
	fmt.Println("✅ Connected to Postgres")

	// make uploads dir if missing : 
	os.MkdirAll("./uploads", os.ModePerm)

	// for applying middlewares : 
	r := mux.NewRouter()

	// public routes :
	r.HandleFunc("/api/signup", handlers.SignupHandler).Methods("POST")
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/api/logout", handlers.LogoutHandler).Methods("POST")
	r.HandleFunc("/api/publicFiles", handlers.PublicFilesHandler).Methods("GET")



	// protected routes with middlewares : 
	// user auth. api on refreshing : 
	r.Handle("/api/me", middleware.AuthMiddleware(
		http.HandlerFunc(handlers.RefershHandler),
	)).Methods("GET")
	
	
	// file upload route : 
	r.Handle("/api/upload", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.UploadHandler)),
		)).Methods("POST")

	// view-files admin route :
	r.Handle("/api/adminFiles", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.AdminFilesHandler)),
		)).Methods("GET")

	// change user roles (admin only)
	r.Handle("/api/makeAdmin", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.MakeAdminHandler)),
		)).Methods("POST")

	r.Handle("/api/makeUser", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.MakeUserHandler)),
		)).Methods("POST")



	// view-files route : 
	r.Handle("/api/files", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.FilesHandler)),
		)).Methods("GET")
	
	// file download route with file_id : 
	r.Handle("/api/fileDownload/{id}", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.FileDownloadHandler)),
		)).Methods("GET")
	
	// file delete route with file_id : 
	r.Handle("/api/fileDelete/{id}", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.FileDeleteHandler)),
		)).Methods("GET")

	// file toggle privacy handler with file_id : 
	r.Handle("/api/fileTogglePrivacy/{id}", middleware.AuthMiddleware(
		middleware.RateLimitMiddleware(http.HandlerFunc(handlers.FileTogglePrivacyHandler)),
		)).Methods("GET")

	// loading the port no. :
	port := config.AppConfig.Port
    if port == "" {port = "8080"}
	
	// running the server :
	fmt.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":"+port, middleware.CORS(r)))
}
