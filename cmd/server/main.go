package main

import (
	"database/sql"
	"fmt"
	"github.com/esojv/jv-eng-backend/internal"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

func main() {
	db, err := sql.Open("sqlite", "blog.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := internal.NewServer(db)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /posts", srv.GetPostsHandler)
	mux.HandleFunc("GET /posts/{slug}", srv.GetPostHandler)

	// Protected routes
	mux.HandleFunc("POST /posts", srv.UpsertPostHandler)
	mux.HandleFunc("PUT /posts/{slug}", srv.UpsertPostHandler)
	mux.HandleFunc("DELETE /posts/{slug}", srv.DeletePostHandler)
	mux.HandleFunc("POST /images", srv.UploadImageHandler)
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))

	fmt.Println("Server running on :8082")
	log.Fatal(http.ListenAndServe(":8082", internal.CorsMiddleware(mux)))
}
