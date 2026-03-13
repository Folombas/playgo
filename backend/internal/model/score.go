package model

import "time"

// Score represents a player score in the leaderboard
type Score struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

// ScoreInput represents the input for creating a new score
type ScoreInput struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}
