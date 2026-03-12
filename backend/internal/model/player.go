package model

// Player represents a game player
type Player struct {
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Score         int     `json:"score"`
	Energy        int     `json:"energy"`
	MaxEnergy     int     `json:"max_energy"`
	TapValue      int     `json:"tap_value"`
	AutoTapPerSec float64 `json:"auto_tap_per_sec"`
	Level         int     `json:"level"`
	XP            int     `json:"xp"`
	XPToNextLevel int     `json:"xp_to_next_level"`
	CreatedAt     int64   `json:"created_at"`
	UpdatedAt     int64   `json:"updated_at"`
}

// PlayerUpgrade represents a player's upgrade
type PlayerUpgrade struct {
	PlayerID string `json:"player_id"`
	UpgradeID string `json:"upgrade_id"`
	Count     int    `json:"count"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	Username   string `json:"username"`
	Score      int    `json:"score"`
	Level      int    `json:"level"`
}

// SaveGameRequest represents a save game request
type SaveGameRequest struct {
	PlayerID      string             `json:"player_id"`
	Username      string             `json:"username"`
	Score         int                `json:"score"`
	Energy        int                `json:"energy"`
	TapValue      int                `json:"tap_value"`
	AutoTapPerSec float64            `json:"auto_tap_per_sec"`
	Level         int                `json:"level"`
	XP            int                `json:"xp"`
	XPToNextLevel int                `json:"xp_to_next_level"`
	Upgrades      []PlayerUpgrade    `json:"upgrades"`
}

// SaveGameResponse represents a save game response
type SaveGameResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetPlayerResponse represents a get player response
type GetPlayerResponse struct {
	Success bool            `json:"success"`
	Player  *Player         `json:"player,omitempty"`
	Upgrades []PlayerUpgrade `json:"upgrades,omitempty"`
}

// LeaderboardResponse represents a leaderboard response
type LeaderboardResponse struct {
	Success bool              `json:"success"`
	Entries []LeaderboardEntry `json:"entries"`
}
