# Go Tap Master Backend

Backend API for Go Tap Master game built with Go.

## API Endpoints

### Save Player Progress
```
POST /api/player/save
Content-Type: application/json

{
  "player_id": "player123",
  "username": "Gopher",
  "score": 1000,
  "energy": 50,
  "tap_value": 2,
  "auto_tap_per_sec": 10.5,
  "level": 5,
  "xp": 50,
  "xp_to_next_level": 100,
  "upgrades": [
    {"upgrade_id": "vars", "count": 5},
    {"upgrade_id": "functions", "count": 2}
  ]
}
```

### Get Player Progress
```
GET /api/player/{id}

Response:
{
  "success": true,
  "player": {
    "id": "player123",
    "username": "Gopher",
    "score": 1000,
    ...
  },
  "upgrades": [...]
}
```

### Get Leaderboard
```
GET /api/leaderboard

Response:
{
  "success": true,
  "entries": [
    {"rank": 1, "username": "Pro", "score": 5000, "level": 10},
    {"rank": 2, "username": "Gopher", "score": 1000, "level": 5}
  ]
}
```

### Health Check
```
GET /health
```

## Running the Server

### Development
```bash
cd backend
go run ./cmd/server
```

### Production Build
```bash
go build -o bin/server ./cmd/server
./bin/server
```

### With Custom Port
```bash
PORT=3000 go run ./cmd/server
```

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── handler/
│   │   └── handler.go       # HTTP handlers
│   ├── model/
│   │   └── player.go        # Data models
│   └── storage/
│       └── memory.go        # In-memory storage
├── go.mod
└── go.sum
```

## Future Enhancements

- [ ] PostgreSQL storage implementation
- [ ] WebSocket for real-time updates
- [ ] Authentication with JWT
- [ ] Redis for session management
- [ ] Docker containerization
