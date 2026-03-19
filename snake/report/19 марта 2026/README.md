# Go79 — 19 марта 2026

## 🎵 Исправлена паника в аудиосистеме — игра зазвучала!

### 📋 Проблема

При запуске игры возникала паника:
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x0 pc=0x7ff646fdc2d0]

goroutine 25 [running]:
github.com/hajimehoshi/ebiten/v2/audio.(*Context).NewPlayer(0x0, ...)
```

**Причина:** В `internal/audio/audio.go` функция `Play()` вызывала `audio.NewPlayerFromBytes(nil, data)` с `nil` контекстом аудио.

### ✅ Решение

#### Изменения в `internal/audio/audio.go`

1. **Добавлено поле контекста в структуру:**
```go
type AudioSystem struct {
    context    *audio.Context  // ← новое поле
    players    map[SoundType][]byte
    volume     float64
    enabled    bool
}
```

2. **Инициализация контекста при создании:**
```go
func NewAudioSystem() *AudioSystem {
    as := &AudioSystem{
        context: audio.NewContext(44100),  // ← sample rate 44100 Гц
        players: make(map[SoundType][]byte),
        volume:  0.3,
        enabled: true,
    }
    // ... предзагрузка звуков
}
```

3. **Использование контекста для создания плеера:**
```go
func (as *AudioSystem) Play(soundType SoundType) {
    // ...
    player := as.context.NewPlayerFromBytes(data)  // ← вместо audio.NewPlayerFromBytes(nil, data)
    player.SetVolume(as.volume)
    player.Rewind()
    player.Play()
}
```

### 🔊 Звуковые эффекты в игре

Аудиосистема генерирует 9 типов звуков программно (без внешних файлов):

| Звук | Событие |
|------|---------|
| 🍎 `SoundEatFood` | Поедание еды |
| 🪙 `SoundCollectCoin` | Сбор монеты |
| 🗝️ `SoundCollectKey` | Сбор ключа |
| 🏴‍☠️ `SoundOpenChest` | Открытие сундука |
| 🏹 `SoundShoot` | Выстрел стрелой |
| 💥 `SoundExplosion` | Взрыв бомбы |
| ⚡ `SoundPowerUp` | Сбор бонуса |
| 💀 `SoundEnemyKill` | Убийство врага |
| ☠️ `SoundGameOver` | Проигрыш |

### 📁 Структура проекта
```
snake/
├── main.go                  # Основной код игры
├── go.mod                   # Go модуль
├── go.sum                   # Зависимости
├── snake.exe                # Скомпилированный бинарник
├── internal/
│   ├── audio/
│   │   └── audio.go         # ✨ Исправленная аудиосистема
│   ├── effects/
│   │   └── effects.go       # Система эффектов
│   ├── game/
│   │   └── game.go          # Игровая логика
│   └── ui/
│       └── renderer.go      # Рендеринг
└── report/
    └── 19 марта 2026/
        └── README.md        # Этот отчёт
```

### 🚀 Запуск игры
```bash
cd D:\Projects\playgo\snake
go run main.go
```

Или запустить скомпилированный бинарник:
```bash
./snake.exe
```

### 🎮 Управление в игре
| Клавиша | Действие |
|---------|----------|
| ↑↓←→ | Движение змейки |
| SPACE | Выстрел стрелой |
| P | Пауза |
| ENTER | Рестарт после Game Over |

### 📌 Коммиты
- `c383413` — Go79: исправить панику в аудиосистеме

### 🎯 Итоги дня
- ✅ Найдена и исправлена критическая ошибка с nil pointer
- ✅ Аудиосистема теперь работает корректно
- ✅ Игра зазвучала эффектными звуками
- ✅ Все 9 типов звуков воспроизводятся при соответствующих событиях

---
**Go365 Challenge** — День 79 из 365 (19 марта 2026)

**Фокус:** Go + Ebitengine = геймдев на Go! 🎮🐍

**Девиз дня:** Звук включён — играем дальше! 🔊🎵
