package handler

import (
	"encoding/json"
	"net/http"
	"playgo/backend/internal/model"
	"playgo/backend/internal/storage"
	"strings"
)

// LeaderboardHandler handles leaderboard HTTP requests
type LeaderboardHandler struct {
	storage *storage.Storage
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(storage *storage.Storage) *LeaderboardHandler {
	return &LeaderboardHandler{storage: storage}
}

// GetScores handles GET /api/leaderboard
func (h *LeaderboardHandler) GetScores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scores, err := h.storage.GetTopScores(10)
	if err != nil {
		http.Error(w, "Failed to get scores", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// CreateScore handles POST /api/leaderboard
func (h *LeaderboardHandler) CreateScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input model.ScoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		input.Name = "Anonymous"
	}
	if input.Score < 0 {
		http.Error(w, "Invalid score", http.StatusBadRequest)
		return
	}

	score, err := h.storage.CreateScore(&input)
	if err != nil {
		http.Error(w, "Failed to create score", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(score)
}
