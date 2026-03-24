# Go365 Day 84 — Отчёт о разработке

**Дата:** 24 марта 2026 года
**Проект:** go_mario v0.9.0
**Go365 День:** 84

---

## 🎮 Версия: go_mario v0.9.0

### Что нового

#### 💥 Комбо-система
Механика подсчёта серий прыжков на врагах с множителями очков:

| Параметр | Значение |
|----------|----------|
| Максимальное комбо | x10 |
| Множитель | 1x + 0.5x за уровень |
| Таймер комбо | 2 секунды (120 кадров) |
| Бонус за MAX комбо | +600 очков за врага |

**Пример расчёта:**
- 1-е комбо: 100 + 150 = 250 очков
- 5-е комбо: 100 + 350 = 450 очков
- 10-е комбо (MAX): 100 + 600 = 700 очков

**Визуальная индикация:**
- 🟡 Золотой (x2-x4): обычное комбо
- 🔴 Красный (x5-x9): высокое комбо
- 🟣 Фиолетовый (x10): MAX комбо

#### 🔴 Пружины (Springs)
Новые объекты уровня для высоких прыжков:

| Характеристика | Описание |
|----------------|----------|
| Шанс появления | 1.5% на tile |
| Цвет мира 1 | 🔴 Красный |
| Цвет мира 2+ | 🟢 Зелёный |
| Цвет мира 4+ | 🔵 Синий |
| Сила отскока | vy = -16 |
| Анимация | 10 кадров сжатия |

**Механика:**
- Игрок прыгает на пружину сверху
- Пружина сжимается (визуальный эффект)
- Игрок получает высокий отскок
- Звуковой эффект при срабатывании

#### 🌀 Телепорты (Portals)
Парные порталы для быстрого перемещения:

| Параметр | Значение |
|----------|----------|
| Цвет портала 1 | 🟣 Фиолетовый (150,50,255) |
| Цвет портала 2 | 🔵 Синий (50,150,255) |
| Появление | Мир 2+, позиция x=50 |
| Расстояние | 30 tiles между порталами |
| Эффект | Мгновенная телепортация |

**Визуальные эффекты:**
- Анимированные swirling-линии
- Эффект свечения (alpha 100)
- Частицы при телепортации (20 частиц)
- Звуковой эффект двери

---

## 📁 Изменённые файлы

### Код
- `go_mario/platformer.go` — основные изменения (+236 строк)
  - Структуры: Spring, Portal
  - Поля Player: combo, comboTimer, lastStompTime
  - Константы: TileSpring, TilePortal, MaxCombo
  - Функции: updateCombo, updateSprings, updatePortals
  - Функции: drawSprings, drawPortals
  - Обновлена drawUI для индикатора комбо

### Документация
- `go_mario/README.md` — обновлён с новыми механиками
  - Добавлены разделы: Комбо-система, Пружины, Телепорты
  - Обновлена структура проекта
  - Добавлено описание врагов и бонусов
  - Версия v0.9.0

### Бинарник
- `go_mario/go_mario.exe` — скомпилированная версия

---

## 🔧 Технические детали

### Комбо-система
```go
// При уничтожении врага
g.player.combo++
if g.player.combo > MaxCombo {
    g.player.combo = MaxCombo
}
g.player.comboTimer = 120 // 2 seconds at 60 FPS

// Множитель очков
comboMultiplier := 1.0 + float64(g.player.combo)*0.5
bonusScore := int(float64(100) * comboMultiplier)
g.player.score += bonusScore
```

### Обновление комбо
```go
func (g *Game) updateCombo() {
    if g.player.combo > 0 {
        g.player.comboTimer--
        if g.player.comboTimer <= 0 {
            g.player.combo = 0
        }
    }
}
```

### Пружины
```go
func (g *Game) updateSprings() {
    for _, spring := range g.level.springs {
        // Проверка коллизии
        if !spring.compressed && g.player.vy > 0 {
            spring.compressed = true
            spring.timer = 10
            g.player.vy = -16 // High bounce!
            g.player.onGround = false
            playSound(SoundPowerup)
        }
    }
}
```

### Телепорты
```go
func (g *Game) updatePortals() {
    for _, portal := range g.level.portals {
        // Проверка коллизии
        if portal.linkedTo != nil {
            g.player.x = portal.linkedTo.x - float64(g.player.width)
            g.player.y = portal.linkedTo.y
            g.player.vx = 0
            g.player.vy = 0
            playSound(SoundDoor)
        }
    }
}
```

### Индикатор комбо в UI
```go
if g.player.combo > 1 {
    comboText := fmt.Sprintf("COMBO\nx%d", g.player.combo)
    comboColor := color.RGBA{255, 215, 0, 255} // Gold
    if g.player.combo >= 5 {
        comboColor = color.RGBA{255, 100, 100, 255} // Red
    }
    if g.player.combo >= MaxCombo {
        comboColor = color.RGBA{150, 50, 255, 255} // Purple
    }
    text.Draw(screen, comboText, gameAssets.gameFont, 420, 28, comboColor)
}
```

---

## 📊 Статистика сборки

```
Строк в коде:     ~2743 (+236)
Файлов изменено:  3
Новых структур:   2 (Spring, Portal)
Новых констант:   3 (TileSpring, TilePortal, MaxCombo)
Зависимостей:     golang.org/x/image v0.31.0
```

---

## ✅ Чеклист Go365

- [x] Ежедневное программирование на Go
- [x] Развитие пет-проекта (go_mario)
- [x] Изучение новых механик (комбо, телепорты)
- [x] Интеграция внешних ресурсов
- [x] Коммит и пуш в репозиторий
- [x] Документирование (PLAN.md, CHANGELOG.md)

---

## 🎯 Фокус

> **Тотальная фокусировка на Go в 2026 году!**

Никакого распыления на C#, Visual Studio, MonoGame. Только Go и Ebitengine.

---

## 📅 Следующий день

**Go85 — 25 марта 2026**

Возможные улучшения:
- Система боссов в конце миров
- Подвижные платформы и лифты
- Сохранение прогресса в файл
- Фоновая музыка

---

**Создано:** 24 марта 2026
**Go365 Challenge:** День 84 из 365
**Версия:** go_mario v0.9.0
