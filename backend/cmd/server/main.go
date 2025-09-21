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
)

func main() {

	// loading environment variables from .env file
	config.LoadConfig()
	
	// connection to postgrSQL : 
	if err := db.Connect(); err != nil {
		log.Fatal("DB connection failed:", err)
	}
	fmt.Println("✅ Connected to Postgres")

	os.MkdirAll("./uploads", os.ModePerm)

	// for applying middlewares : 
	mux := http.NewServeMux()

	// public routes :
	mux.HandleFunc("/api/signup", handlers.SignupHandler)
	mux.HandleFunc("/api/login", handlers.LoginHandler)
	mux.HandleFunc("/api/logout",handlers.LogoutHandler)

	// protected routes  with applied middlewares and JWT integration :
	mux.Handle("/api/me", middleware.AuthMiddleware(http.HandlerFunc(handlers.RefershHandler)))
	mux.Handle("/api/upload", middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadHandler)))
	mux.Handle("/api/files", middleware.AuthMiddleware(http.HandlerFunc(handlers.FilesHandler)))

	// loading the port no. :
	port := os.Getenv("PORT")
    if port == "" {port = "8080"}
	
	// running the server :
	fmt.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":"+port, middleware.CORS(mux)))
}
