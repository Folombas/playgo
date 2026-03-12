# Go Tap Master 🐹

A Hamster Combat-style tap game about learning Go programming, built with **TypeScript + React + Phaser 3 + Webpack**.

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

## Tech Stack

- **TypeScript** - Type-safe JavaScript
- **React 19** - UI framework
- **Phaser 3** - Game framework
- **Webpack 5** - Module bundler

## Getting Started

### Install dependencies

```bash
npm install
```

### Development mode

```bash
npm run dev
```

This will start the webpack dev server at `http://localhost:8080` with hot reload.

### Production build

```bash
npm run build
```

The built files will be in the `dist/` directory.

## Project Structure

```
playgo/
├── public/
│   └── index.html          # HTML template
├── src/
│   ├── components/
│   │   └── App.tsx         # Main React component
│   ├── game/
│   │   ├── GameScene.ts    # Phaser game scene
│   │   └── types.ts        # TypeScript types
│   ├── index.tsx           # Entry point
│   └── styles.css          # Styles
├── package.json
├── tsconfig.json
└── webpack.config.js
```

## How to Play

1. Run `npm run dev` to start the development server
2. Open `http://localhost:8080` in your browser
3. Click/tap the Gopher to earn coins
4. Buy upgrades to earn coins automatically
5. Level up to increase your tap value
6. Learn Go facts displayed at the bottom

## Controls

- **Click/Tap** - Tap the Gopher to earn coins
- **Upgrades button** - Open/close the upgrades panel
- **BUY** - Purchase upgrades to increase auto-income

## License

MIT
