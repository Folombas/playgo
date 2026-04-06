# Go365 Day 98 - Daily Report
## Date: 6 апреля 2026 (понедельник)

### Project: Dungeon Crawler

---

## 📋 Summary

Сегодня полностью переосмыслил подход к разработке игр на Go! Вместо бесконечного пересоздания платформеров, создал **принципиально новый жанр** — **Dungeon Crawler с roguelike-элементами** на Ebitengine.

Это ** roguelike-приключение** с процедурной генерацией подземелий, 4 типами врагов с разным AI, системой предметов, боевой механикой и 10 этажами для прохождения!

---

## 🏗 Architecture

### Модульная структура проекта:
```
Go98/
├── cmd/game/main.go           # Точка входа, Game struct, основной цикл
├── internal/
│   ├── config/config.go       # Константы, настройки управления
│   ├── audio/audio.go         # Процедурные звуки (12 эффектов)
│   ├── dungeon/dungeon.go     # Процедурная генерация подземелий
│   ├── entity/
│   │   ├── entity.go          # Базовый интерфейс + коллизии
│   │   ├── player.go          # Игрок (HP, атака, инвентарь)
│   │   ├── enemy.go           # 4 типа врагов с AI
│   │   └── item.go            # Предметы (монеты, зелья, ключи)
│   ├── engine/
│   │   └── renderer.go        # Камера, рендеринг dungeon
│   ├── helper/draw.go         # Утилита для рисования прямоугольников
│   ├── sprite/sprite.go       # Загрузка и кэширование спрайтов
│   └── ui/hud.go              # HUD, меню, пауза, Game Over, Victory
├── assets/sprites/            # 300+ спрайтов из коллекции
│   ├── tiles/                 # Тайлы стен/полов
│   ├── items/                 # Монеты, ключи, зелья, кристаллы
│   ├── player/                # Спрайты персонажа
│   ├── enemies/               # Спрайты врагов
│   └── hud/                   # Элементы интерфейса
└── report/daily-report.md     # Этот файл
```

---

## 🎮 Game Features

### Procedural Dungeon Generation
- **Room-based генерация**: 4-10 комнат случайного размера (4x4 - 8x8 тайлов)
- **L-образные коридоры**: Соединяют все комнаты
- **Дополнительные соединения**: Создают петли для альтернативных путей
- **Прогрессия сложности**: Больше врагов, шипов и ловушек на глубоких этажах
- **10 этажей**: Каждый уникален, нужно добраться до 10-го

### Enemies (4 типа с разным AI)
1. **Slime (Зелёный)**: Патрулирует территорию, движется туда-сюда
2. **Bee (Жёлтый)**: Летает, преследует игрока только в зоне видимости
3. **Fly (Красный)**: Агрессивный, всегда преследует игрока
4. **Snail (Коричневый)**: Стационарный, наносит урон при контакте

**Прогрессия врагов**: HP и урон врагов растут с каждым этажом (+5 HP, +2-3 DMG)

### Combat System
- **Направление атаки**: Зависит от направления взгляда игрока (WASD)
- **Attack Hitbox**: 24x24 пикселя перед игроком
- **Кулдаун атаки**: 20 кадров между атаками
- **Active frame**: Урон наносится на 10-м кадре анимации
- **Визуальный эффект**: Жёлтый квадрат при атаке
- **I-frames**: 60 кадров неуязвимости после получения урона
- **Floating damage**: Числа урона поднимаются вверх

### Items & Loot
- **Coins (Золотые)**: +10 очков, собираются автоматически
- **Gems (Синие)**: +50 очков, редкие
- **Keys (Жёлтые)**: Для будущих дверей (пока декоративные)
- **Potions (Розовые)**: Восстанавливают 30 HP, используются клавишей K
- **Random Chests**: Будущее расширение

### Tile Types
- **Floor (Серый)**: Проходимый
- **Wall (Тёмно-серый)**: Непроходимый, с затенением
- **Door (Коричневый)**: Пока декоративный
- **Stairs (Жёлтый)**: Переход на следующий этаж
- **Spikes (Шипы)**: Наносят 5 урона при наступании
- **Water (Синий)**: Непроходимая ловушка

### Player Stats
- **HP**: 100 (с цветным HP-баром)
- **Attack Damage**: 10 (+10% за убийство)
- **Speed**: 3 пикселя/кадр
- **Inventory**: Coins, Gems, Keys, Potions
- **Score**: Очки за убийства + сбор предметов + бонусы за этажи

### Screens
- **Main Menu**: 3 опции (Start, How to Play, Exit), навигация W/S
- **Playing**: Полный геймплей с HUD
- **Pause**: Overlay с подсказкой
- **Game Over**: Финальный счёт + достигнутый этаж
- **Victory**: Экран победы после прохождения 10 этажей
- **Floor Transition**: Анимация перехода между этажами

---

## 🔧 Technologies & Patterns

### Go Concepts Used
- **Interfaces**: Entity interface для统一ного управления сущностями
- **Composition**: Base struct + специфичные поля
- **Packages**: Модульная архитектура (9 пакетов)
- **Error Handling**: Консистентная обработка ошибок
- **Concurrency Safety**: sync.Mutex в audio менеджере

### Ebitengine Features
- **Drawing**: ebiten.DrawImage, масштабирование 1x1 изображения для прямоугольников
- **Input**: inpututil.IsKeyJustPressed для отзывчивого управления
- **Audio**: audio.Context + NewPlayerFromBytes для процедурных звуков
- **Game Loop**: ebiten.Game interface (Update, Draw, Layout)
- **Window**: Resizing, custom title

### Design Patterns
- **Entity-Component inspired**: Базовый компонент + специализация
- **Manager pattern**: SpriteManager, SoundManager
- **State Machine**: GameState enum для экранов
- **Factory functions**: NewPlayer, NewEnemy, NewItem, NewDungeon
- **Camera follow**: Smooth lerp (0.1 factor)

### Algorithms
- **Random Dungeon Generation**: Room placement с overlap checking
- **L-shaped Corridors**: Случайный выбор horizontal-first или vertical-first
- **AABB Collision**: Проверка пересечения прямоугольников
- **Procedural Audio**: Square wave generation с binary encoding

---

## 📊 Stats

- **Lines of Code**: ~1800+ (12 файлов Go)
- **Packages**: 9 (config, audio, dungeon, entity×4, engine, helper, sprite, ui)
- **Files**: 12 Go files + 300+ спрайтов
- **Game States**: 6 (Menu, Playing, Paused, GameOver, Victory, NextFloor)
- **Enemy Types**: 4 с уникальным AI
- **Item Types**: 5 (Coin, Gem, Key, Potion, Chest)
- **Sound Effects**: 12 процедурных
- **Development Time**: ~4 часа

---

## 🎯 Learning Outcomes

### Что освоил сегодня:
1. ✅ **Процедурная генерация** — room-based dungeon с коридорами
2. ✅ **AI поведения врагов** — patrol, alert, chase patterns
3. ✅ **Боевая система** — directional attacks, hitboxes, i-frames
4. ✅ **Система предметов** — pickup, inventory, usage
5. ✅ **Процедурные звуки** — square wave, audio context
6. ✅ **Модульная архитектура** — 9 пакетов, interfaces
7. ✅ **Ebitengine v2.9 API** — правильный DrawImage подход
8. ✅ **Камера** — smooth follow с boundaries
9. ✅ **Game States** — state machine с переходами
10. ✅ **Прогрессия** — 10 этажей с scaling difficulty

### Go Concepts:
- Interface design и когда их использовать
- Package organization и избегание циклических импортов
- Composition over inheritance (Base struct)
- Error handling patterns
- Concurrency safety (sync.Mutex)

---

## 🚀 How to Run

```bash
cd D:\Projects\playgo\Go98
go run ./cmd/game
```

Или запустить скомпилированный файл:
```bash
.\dungeon_crawler.exe
```

### Controls:
- **W/A/S/D** — Движение
- **J** — Атака (в направлении взгляда)
- **K** — Использовать зелье
- **ESC** — Пауза
- **ENTER** — Выбрать в меню / Рестарт

---

## 🎮 Gameplay Loop

```
Start Game → Spawn на 1-м этаже
    → Исследуй комнаты, собирай предметы
    → Сражайся с врагами (получай очки)
    → Избегай шипов и ловушек
    → Найди Stairs → Next Floor (+100 очков)
    → ... повторить 10 раз ...
    → VICTORY! (Бонус +1000 очков)
```

**Risk vs Reward**: Глубже = сложнее враги, но больше очков и сокровищ!

---

## 🔮 Future Improvements

Потенциальные расширения (завтра или позже):
- [ ] Двери и ключи (locked rooms)
- [ ] Босс на 10-м этаже
- [ ] Больше типов врагов (5+)
- [ ] Система улучшений оружия
- [ ] Мини-карта
- [ ] High Score таблица
- [ ] Сохранение прогресса
- [ ] Реальные PNG спрайты вместо прямоугольников
- [ ] Particle effects (взрывы, магия)
- [ ] Больше звуков (ambient, music)

---

## 💭 Reflections

Сегодняшний день показал рост как разработчика:
1. **Научился** — procedural generation, AI behaviors, game architecture
2. **Понял** — важность модульности (легко тестировать пакеты отдельно)
3. **Избежал** — циклических импортов через helper пакет
4. **Практиковал** — Ebitengine API, Go interfaces, composition

**Главный вывод:** Лучше сделать одну хорошую игру с глубиной, чем 10 поверхностных! Dungeon Crawler — это первый проект с **настоящей реиграбельностью** (каждый проход уникален).

---

*Go365 Day 98 ✅ — Продолжаю путь к Go-разработчику!* 🚀
