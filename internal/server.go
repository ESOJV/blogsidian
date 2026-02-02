package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := GetPosts(s.DB)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (s *Server) UpsertPostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	post, err := ParsePost(body)
	if err != nil {
		http.Error(w, "Failed to parse post", http.StatusBadRequest)
		return
	}

	requestSlug := r.PathValue("slug")
	if requestSlug != "" && requestSlug != post.Slug {
		http.Error(w, "Request slug and frontmatter slug must be the same", http.StatusBadRequest)
		return
	}

	err = UpsertPost(s.DB, post)
	if err != nil {
		http.Error(w, "Failed to save post", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPut {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Post updated successfully"))
	}

	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Post created successfully"))
	}
}

func (s *Server) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	post, err := GetPostBySlug(s.DB, slug)
	if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	if post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (s *Server) DeletePostHandler(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	err := DeletePostBySlug(s.DB, slug)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Post deleted successfully"))
}

func (s *Server) UploadImageHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error Parsing Multi Part form", http.StatusBadRequest)
		return
	}

	formFile, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error Parsing Multi Part form", http.StatusBadRequest)
		return
	}
	defer formFile.Close()

	err = os.MkdirAll("images", 0755)
	if err != nil {
		http.Error(w, "Error creating server directory", http.StatusInternalServerError)
		return
	}

	serverFile, err := os.Create("images/" + header.Filename)
	if err != nil {
		http.Error(w, "Error creating file on server", http.StatusInternalServerError)
		return
	}
	defer serverFile.Close()

	_, err = io.Copy(serverFile, formFile)
	if err != nil {
		http.Error(w, "Error Copying file to server", http.StatusInternalServerError)
		return
	}
	filePath := fmt.Sprintf("filepath: %s", serverFile.Name())

	w.Write([]byte(filePath))

}
