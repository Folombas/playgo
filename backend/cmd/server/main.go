package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"playgo/backend/internal/handler"
	"playgo/backend/internal/storage"
)

func main() {
	// Initialize storage
	dbPath := "./playgo.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create handlers
	leaderboardHandler := handler.NewLeaderboardHandler(store)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json"))

	// CORS (for development)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Get("/api/leaderboard", leaderboardHandler.GetScores)
	r.Post("/api/leaderboard", leaderboardHandler.CreateScore)

	// Static files (frontend)
	fs := http.FileServer(http.Dir("../frontend"))
	r.Handle("/*", fs)

	// Server
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	addr := fmt.Sprintf(":%s", port)

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		log.Println("🛑 Shutting down server...")
		store.Close()
		os.Exit(0)
	}()

	log.Printf("🚀 Server starting on http://localhost%s\n", addr)
	log.Println("📊 Leaderboard API: http://localhost%s/api/leaderboard", addr)
	log.Println("🎮 Game: http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
