# Go365 Day 86 — Отчёт о разработке

**Дата:** 26 марта 2026 года
**Проект:** go_mario v2.0.0 (Roguelite Survivor)
**Go365 День:** 86

---

## 🎮 ПЕРЕРАБОТКА: go_mario v2.0.0

### Новая концепция
**Roguelite Survivor** в стиле Vampire Survivors!

### Что сделано

#### 1. Геймплей
- [x] Авто-атака ближайшего врага
- [x] 4 типа оружия (Magic Missile, Fireball, Lightning, Aura)
- [x] Система волн (каждые 30 секунд)
- [x] Спавн врагов вокруг игрока
- [x] Подбор опыта (магнитный эффект)

#### 2. Прокачка
- [x] Система уровней (EXP → Level Up)
- [x] Выбор из 3 случайных апгрейдов
- [x] 6 типов апгрейдов (урон, скорость, HP, размер, проникновение, новое оружие)
- [x] Применение апгрейдов

#### 3. Враги
- [x] 3 типа врагов (Basic, Fast, Tank)
- [x] ИИ преследования игрока
- [x] Разделение между врагами (separation)
- [x] Прогрессия сложности по волнам

#### 4. Визуальные эффекты
- [x] Система частиц (смерть врагов)
- [x] Всплывающий урон
- [x] Полоски здоровья врагов
- [x] Анимация игрока (ходьба)
- [x] Мигание при неуязвимости

#### 5. Интерфейс
- [x] HUD: HP, EXP, уровень, золото
- [x] Индикаторы кулдауна оружия
- [x] Таймер выживания
- [x] Счётчик убийств и волн
- [x] Overlay выбора апгрейдов
- [x] Меню и Game Over экран

#### 6. Аудио
- [ ] Звуковые эффекты (отложено)
- [ ] Фоновая музыка (отложено)

---

## 📁 Изменённые файлы

### Код
- `go_mario/main.go` — полностью переписан (~1150 строк)

### Документация
- `go_mario/README.md` — обновлён с новой концепцией

### Бинарник
- `go_mario/go_mario.exe` — скомпилированная версия

---

## 📊 Статистика

```
Написано строк:     ~1150
Структур:           10+ (Player, Enemy, Projectile, Weapon, Pickup, Particle, DamageNumber, Game...)
Типов оружия:       4
Типов апгрейдов:    6
Типов врагов:       3
Функций:            40+
```

---

## 🎮 Геймплей

### Механики
1. **Движение:** WASD / Стрелки
2. **Авто-атака:** Оружие стреляет автоматически
3. **Сбор опыта:** Подбирай кристаллы с врагов
4. **Левел-ап:** Выбирай 1 из 3 апгрейдов (клавиши 1/2/3)
5. **Выживание:** Держись как можно дольше!

### Прогрессия
- Волна 1: 8 врагов (Basic)
- Волна 2: 11 врагов (Basic + Fast)
- Волна 3+: 14+ врагов (все типы)
- Каждые 30 секунд новая волна

---

## 🔧 Технические детали

### Авто-атака
```go
// Поиск ближайшего врага
var target *Enemy
minDist := float64(500)
for _, e := range g.enemies {
    dist := math.Hypot(e.x-p.x, e.y-p.y)
    if dist < minDist {
        minDist = dist
        target = e
    }
}

// Выстрел снаряда
angle := math.Atan2(target.y-p.y, target.x-p.x)
g.projectiles = append(g.projectiles, &Projectile{
    vx: math.Cos(angle) * speed,
    vy: math.Sin(angle) * speed,
    damage: int(float64(weapon.damage) * p.damageMult),
    pierce: weapon.pierce,
})
```

### Система апгрейдов
```go
type Upgrade struct {
    id          UpgradeType
    name        string
    description string
    icon        string
    tier        int
}

func (g *Game) applyUpgrade(upgrade Upgrade) {
    switch upgrade.id {
    case UpgradeDamage:
        p.damageMult += 0.2
    case UpgradeAttackSpeed:
        p.attackSpeed += 0.15
    case UpgradeMaxHealth:
        p.maxHealth += 20
        p.health += 20
    // ...
    }
}
```

### ИИ врагов
```go
// Движение к игроку
angle := math.Atan2(p.y-e.y, p.x-e.x)
e.vx = math.Cos(angle) * e.speed
e.vy = math.Sin(angle) * e.speed

// Разделение (separation)
for _, other := range g.enemies {
    dist := math.Hypot(e.x-other.x, e.y-other.y)
    if dist < 30 {
        pushAngle := math.Atan2(e.y-other.y, e.x-other.x)
        e.vx += math.Cos(pushAngle) * 0.5
    }
}
```

---

## ✅ Чеклист Go365

- [x] Ежедневное программирование на Go
- [x] Кардинальная смена концепции (платформер → roguelite)
- [x] Использование всей мощи Ebitengine
- [x] Коммит и пуш в репозиторий
- [x] Документирование (README.md, CHANGELOG.md)
- [x] Сборка без ошибок

---

## 🎯 Фокус

> **Тотальная фокусировка на Go в 2026 году!**

Никакого распыления. Только Go и Ebitengine.

---

## 📅 Следующий день

**Go87 — 27 марта 2026**

Планы:
- Добавить боссов (каждые 5 волн)
- Новые типы оружия
- Пассивные предметы
- Баланс и полировка

---

**Создано:** 26 марта 2026
**Go365 Challenge:** День 86 из 365
**Версия:** go_mario v2.0.0 (Roguelite Survivor)
