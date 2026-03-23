# Go365 Day 83 — Отчёт о разработке

**Дата:** 23 марта 2026 года  
**Проект:** go_mario v0.6.0  
**Go365 День:** 83

---

## 🎮 Версия: go_mario v0.6.0

### Что нового

#### 🎨 Пользовательский шрифт UI
- **Шрифт:** SuperAdorable (24pt, 72 DPI)
- **Загрузка:** golang.org/x/image/font/opentype
- **Применение:**
  - Главное меню (заголовок, инструкции, версия)
  - HUD (счёт, монеты, мир, жизни)
  - Экран Game Over
  - Экран Course Clear

#### 🔊 Звуковые эффекты Kenney RPG Audio
Интегрировано 13 звуковых файлов:

| Звук | Файл | Описание |
|------|------|----------|
| Jump | footstep00.ogg | Прыжок игрока |
| Footstep | footstep01.ogg | Шаги при движении |
| Coin | handleCoins.ogg | Сбор монет |
| Stomp | cloth1.ogg | Уничтожение врага |
| Hit | creak1.ogg | Получение урона |
| Die | doorClose_1.ogg | Смерть игрока |
| Powerup | bookOpen.ogg | Получение бонуса |
| Bump | bookPlace1.ogg | Удар о блок |
| Break | knifeSlice.ogg | Разрушение блока |
| Start | metalClick.ogg | Начало игры |
| Win | handleCoins2.ogg | Победа на уровне |
| Door | doorOpen_1.ogg | Двери/трубы |
| Item | beltHandle1.ogg | Предметы |

**Fallback:** Процедурные звуки (генерация синусоиды) если файлы не найдены

#### 🖼️ Визуальные улучшения

**Главное меню:**
- Градиентный фон (синий, от тёмного к светлому)
- Тень заголовка "SUPER GO MARIO"
- Золотой заголовок с тенью
- Декоративная белая линия
- Цветовая кодировка инструкций
- Версия Go365 внизу экрана

**HUD (верхняя панель):**
- Тёмная полупрозрачная панель (180/255)
- Разделительная линия
- Счёт: белый цвет
- Монеты: золотой цвет
- Мир: зелёный цвет
- Жизни: красный цвет

**Game Over:**
- Тёмно-красный фон
- Красный заголовок
- Финальный счёт
- Призыв к рестарту

**Course Clear:**
- Победный градиент (зелёно-голубой)
- Золотой заголовок
- Счёт и собранные монеты
- Призыв к продолжению

---

## 📁 Изменённые файлы

### Код
- `go_mario/platformer.go` — основные изменения (+219 строк)
- `go_mario/go.mod` — добавлена зависимость golang.org/x/image
- `go_mario/go.sum` — обновлены зависимости

### Ресурсы (новые)
- `go_mario/assets/fonts/` — шрифты (SuperAdorable-MAvyp.ttf)
- `go_mario/assets/sounds/` — звуки Kenney RPG Audio (51 файл)

### Ресурсы (репозиторий)
- `fonts/` — коллекция шрифтов (8 zip-архивов)
- `sounds/` — коллекция звуков Kenney (6 zip-архивов)
- `sprites/` — коллекция спрайтов (21 файл)
- `UI/` — UI-ассеты (5 zip-архивов)

---

## 🔧 Технические детали

### Загрузка шрифтов
```go
func loadFont(path string, size int) (font.Face, error) {
    data, err := os.ReadFile(path)
    ttFont, err := opentype.Parse(data)
    return opentype.NewFace(ttFont, &opentype.FaceOptions{
        Size: float64(size),
        DPI:  72,
    })
}
```

### Загрузка звуков
```go
func playSound(sound SoundType) {
    filePath, ok := soundFiles[sound]
    if ok {
        data, err := os.ReadFile("assets/sounds/" + filePath)
        player := audioCtx.NewPlayerFromBytes(data)
        player.SetVolume(0.5)
        player.Play()
        return
    }
    // Fallback к процедурному звуку
}
```

### Звуки шагов
```go
if p.onGround && p.animFrame%20 == 0 {
    playSound(SoundFootstep)
}
```

---

## 📊 Статистика сборки

```
Строк в коде:     1581 (+219)
Файлов изменено:  100
Звуков:           13
Шрифтов:          1
Зависимостей:     golang.org/x/image v0.31.0
```

---

## ✅ Чеклист Go365

- [x] Ежедневное программирование на Go
- [x] Развитие пет-проекта (go_mario)
- [x] Изучение новых библиотек (font/opentype)
- [x] Интеграция внешних ресурсов
- [x] Коммит и пуш в репозиторий
- [x] Документирование (PLAN.md, CHANGELOG.md)

---

## 🎯 Фокус

> **Тотальная фокусировка на Go в 2026 году!**

Никакого распыления на C#, Visual Studio, MonoGame. Только Go и Ebitengine.

---

## 📅 Следующий день

**Go84 — 24 марта 2026**

Планы:
- Продолжение разработки go_mario
- Возможные улучшения: новые уровни, боссы, анимации
- Оптимизация производительности

---

**Создано:** 23 марта 2026  
**Go365 Challenge:** День 83 из 365
