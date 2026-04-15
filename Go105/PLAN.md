# 🎮 PLAN: Geometric Match Game (Go + Ebitengine)

**Дата:** 15 апреля 2026  
**День:** Go105 (Go365 Challenge)  
**Цель:** Создать современную казуальную игру с геометрическими фигурами

## Описание игры
Казуальная «карманная» игра, где геометрические фигуры (блоки) падают сверху, и игрок должен ставить их друг к другу так, чтобы они сходились — образовывали заполненные линии, которые исчезают и приносят очки.

## Технические требования
- **Язык:** Go
- **Движок:** Ebitengine v2.6+
- **Платформы:** Windows, Android (gomobile), Web (WASM)
- **Организация кода:** модульная, несколько файлов

## Архитектура файлов

| Файл | Назначение |
|------|------------|
| `main.go` | Точка входа, инициализация Ebitengine |
| `game.go` | Основная игровая логика, состояния |
| `board.go` | Игровое поле, проверка линий, коллизии |
| `tetromino.go` | Фигуры (тетрамино), повороты, цвета |
| `input.go` | Обработка клавиатуры, мыши, тача |
| `ui.go` | UI элементы: счёт, меню, Game Over |
| `animation.go` | Анимации: падение, исчезновение, вспышки |
| `assets.go` | Загрузка/создание спрайтов, шрифты |
| `sounds.go` | Звуковые эффекты |

## Игровые состояния
1. `Menu` — главное меню
2. `Playing` — игра
3. `Paused` — пауза
4. `GameOver` — конец игры

## Механики
- Падение фигур с ускорением
- Поворот, перемещение, мгновенная установка
- Удаление заполненных линий
- Бонусные очки за «идеальную установку»
- Прогрессия сложности (каждые 10 линий)
- Сохранение рекорда (JSON)

## Визуальный стиль
- Минимализм, закруглённые углы
- Градиенты на фигурах
- Тень-превью места падения
- Анимированный фон (звёзды/градиент)
- Вспышка при удалении линии

## План коммитов (10+)
1. `init: project structure and go.mod`
2. `feat: main.go entry point with Ebitengine setup`
3. `feat: board.go - game grid and collision detection`
4. `feat: tetromino.go - 7 tetromino types with rotations`
5. `feat: game.go - core game loop and states`
6. `feat: input.go - keyboard and mouse handling`
7. `feat: ui.go - score, level, next figure preview`
8. `feat: animation.go - line clear, drop shadow, effects`
9. `feat: assets.go and sounds.go - procedural assets`
10. `feat: high score, game over screen, polish`
11. `docs: PLAN.md, CHANGELOG.md, build instructions`
12. `build: web assembly support and final testing`

## Сборка
```bash
# Windows
go build -o matchgame.exe

# Web (WASM)
GOOS=js GOARCH=wasm go build -o matchgame.wasm

# Android
gomobile build -target=android
```
