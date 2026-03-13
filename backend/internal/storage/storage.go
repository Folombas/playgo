package storage

import (
	"database/sql"
	"log"
	"playgo/backend/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

// Storage handles database operations
type Storage struct {
	db *sql.DB
}

// New creates a new storage instance
func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Storage{db: db}
	if err := s.initDB(); err != nil {
		return nil, err
	}

	return s, nil
}

// initDB creates the necessary tables
func (s *Storage) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		score INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_score ON scores(score DESC);
	`
	_, err := s.db.Exec(query)
	return err
}

// CreateScore adds a new score to the database
func (s *Storage) CreateScore(score *model.ScoreInput) (*model.Score, error) {
	result, err := s.db.Exec(
		"INSERT INTO scores (name, score) VALUES (?, ?)",
		score.Name, score.Score,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.Score{
		ID:    id,
		Name:  score.Name,
		Score: score.Score,
	}, nil
}

// GetTopScores returns the top scores
func (s *Storage) GetTopScores(limit int) ([]model.Score, error) {
	rows, err := s.db.Query(
		"SELECT id, name, score, created_at FROM scores ORDER BY score DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make([]model.Score, 0)
	for rows.Next() {
		var score model.Score
		err := rows.Scan(&score.ID, &score.Name, &score.Score, &score.CreatedAt)
		if err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}

	return scores, rows.Err()
}

// Close closes the database connection
func (s *Storage) Close() error {
	log.Println("🛑 Closing database connection...")
	return s.db.Close()
}
