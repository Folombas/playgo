package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/playgo/backend/internal/handler"
	"github.com/playgo/backend/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize handlers
	h := handler.NewHandler(store)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handler.CORS)

	// API routes
	r.Mount("/", h.Routes())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("🚀 Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
