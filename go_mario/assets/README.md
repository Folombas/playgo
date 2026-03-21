# 🎨 Как установить спрайты для Go Mario

## 📦 Шаг 1: Скачай спрайт-паки

### Вариант A: Platformer Pack (рекомендую)
**Ссылка:** https://opengameart.org/content/platformer-pack-redux-360-assets

1. Скачай ZIP архив
2. Распакуй
3. Скопируй файлы в папку `assets/`

### Вариант B: Mario-style
**Ссылка:** https://opengameart.org/content/super-mario-bros-sprites

1. Скачай спрайты
2. Положи в `assets/sprites/`

## 📁 Шаг 2: Структура файлов

```
go_mario/
├── assets/
│   ├── sprites/
│   │   ├── player_idle.png    # Игрок (стоит)
│   │   ├── player_run.png     # Игрок (бежит, 8 кадров)
│   │   ├── player_jump.png    # Игрок (прыгает)
│   │   ├── enemy.png          # Враг (Goomba)
│   │   ├── coin.png           # Монета
│   │   ├── mushroom.png       # Гриб
│   │   └── star.png           # Звезда
│   └── tiles/
│       ├── grass.png          # Трава 40x40
│       ├── dirt.png           # Земля 40x40
│       ├── stone.png          # Камень 40x40
│       ├── brick.png          # Кирпичи 40x40
│       └── wood.png           # Дерево 40x40
├── main.go
└── go_mario.exe
```

## 🎮 Шаг 3: Запусти игру

```powershell
go run .
```

## ⚠️ Если спрайтов нет

Игра будет использовать **улучшенную процедурную графику** (не квадраты!).

---

## 🔗 Полезные ресурсы:

- **Kenney.nl** — https://kenney.nl/assets/platformer-pack
- **itch.io Game Assets** — https://itch.io/game-assets
- **OpenGameArt** — https://opengameart.org/
