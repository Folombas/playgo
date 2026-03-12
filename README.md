# Go Tap Master 🐹

A Hamster Combat-style tap game about learning Go programming, built with **TypeScript + React + Phaser 3 + Webpack** and **Go backend**.

## Features

- **Tap to earn** GopherCoins by clicking the Go Gopher mascot
- **Energy system** - manage your energy while tapping
- **6 upgrade types** representing Go concepts:
  - 📦 Variables
  - ⚡ Functions
  - 🏗️ Structs
  - 🔌 Interfaces
  - 🔄 Goroutines
  - 📡 Channels
- **Level system** with XP progression
- **Auto-income** from purchased upgrades
- **Go facts** - learn Go programming while playing!
- **Cloud save** - save progress to Go backend
- **Leaderboard** - compete with other players

## Tech Stack

### Frontend
- **TypeScript** - Type-safe JavaScript
- **React 19** - UI framework
- **Phaser 3** - Game framework
- **Webpack 5** - Module bundler

### Backend
- **Go 1.21** - Backend API
- **chi** - Lightweight HTTP router
- **In-memory storage** - For demo (PostgreSQL coming soon)

## Getting Started

### Frontend

```bash
cd playgo
npm install
npm run dev
```

Open http://localhost:8080

### Backend

```bash
cd backend
go run ./cmd/server
```

API available at http://localhost:8081

## Project Structure

```
playgo/
├── frontend/                    # React + Phaser frontend
│   ├── public/
│   ├── src/
│   │   ├── components/
│   │   ├── game/
│   │   └── index.tsx
│   ├── package.json
│   └── webpack.config.js
├── backend/                     # Go backend
│   ├── cmd/
│   │   └── server/
│   ├── internal/
│   │   ├── handler/
│   │   ├── model/
│   │   └── storage/
│   └── go.mod
└── README.md
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/player/save` | Save player progress |
| GET | `/api/player/{id}` | Get player data |
| GET | `/api/leaderboard` | Get top players |
| GET | `/health` | Health check |

## Development Commands

### Frontend
```bash
npm run dev      # Start dev server
npm run build    # Production build
```

### Backend
```bash
go run ./cmd/server              # Run server
go build -o bin/server ./cmd/server  # Build binary
```

## Roadmap

- [ ] PostgreSQL integration
- [ ] User authentication
- [ ] Real-time multiplayer
- [ ] Mobile app (React Native)
- [ ] More Go concepts
- [ ] Achievements system

## License

MIT
