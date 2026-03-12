package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/playgo/backend/internal/model"
	"github.com/playgo/backend/internal/storage"
)

// Handler handles HTTP requests
type Handler struct {
	storage storage.Storage
}

// NewHandler creates a new handler
func NewHandler(storage storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// Routes sets up HTTP routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/api/player/save", h.SavePlayer)
	r.Get("/api/player/{id}", h.GetPlayer)
	r.Get("/api/leaderboard", h.GetLeaderboard)
	return r
}

// SavePlayer handles saving player progress
func (h *Handler) SavePlayer(w http.ResponseWriter, r *http.Request) {
	var req model.SaveGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	player := &model.Player{
		ID:            req.PlayerID,
		Username:      req.Username,
		Score:         req.Score,
		Energy:        req.Energy,
		MaxEnergy:     100,
		TapValue:      req.TapValue,
		AutoTapPerSec: req.AutoTapPerSec,
		Level:         req.Level,
		XP:            req.XP,
		XPToNextLevel: req.XPToNextLevel,
	}

	if err := h.storage.SavePlayer(player); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save upgrades
	for _, upgrade := range req.Upgrades {
		upgrade.PlayerID = req.PlayerID
		if err := h.storage.SaveUpgrade(&upgrade); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp := model.SaveGameResponse{
		Success: true,
		Message: "Game saved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPlayer handles getting player progress
func (h *Handler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Player ID required", http.StatusBadRequest)
		return
	}

	player, err := h.storage.GetPlayer(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	upgrades, err := h.storage.GetUpgrades(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := model.GetPlayerResponse{
		Success:  true,
		Player:   player,
		Upgrades: upgrades,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLeaderboard handles getting leaderboard
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 10
	entries, err := h.storage.GetLeaderboard(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := model.LeaderboardResponse{
		Success: true,
		Entries: entries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
