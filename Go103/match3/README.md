# 💎 Crystal Cascade - Match-3 Game

Казуальная игра "Три в ряд" на движке Ebitengine. Создана в рамках челленджа Go365 (День 103).

## 🎮 Описание

Crystal Cascade - это классическая Match-3 игра с дружелюбным интерфейсом и увлекательным геймплеем:

- **Игровое поле**: сетка 8x8 с 6 типами кристаллов
- **Механика**: меняйте местами соседние кристаллы, чтобы собрать 3+ в ряд
- **Каскады**: комбо-система с множителями очков
- **Таймер**: 60 секунд на партию
- **Подсказки**: автоматические подсказки при бездействии
- **Бонусы**: 50 очков за 4 в ряд, 100 за 5+ в ряд

## 🚀 Установка и запуск

### Требования

- Go 1.21+
- Ebitengine v2.9.9

### Запуск на вашей системе

```bash
cd match3
go run ./cmd
```

### Сборка

#### Windows
```bash
cd match3
GOOS=windows GOARCH=amd64 go build -o crystal-cascade.exe ./cmd
```

#### Linux
```bash
cd match3
go build -o crystal-cascade ./cmd
```

#### macOS
```bash
cd match3
GOOS=darwin GOARCH=amd64 go build -o crystal-cascade ./cmd
```

#### Android (через gomobile)
```bash
# Установка gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Сборка APK
cd match3
gomobile build -target=android -o crystal-cascade.apk ./cmd
```

#### Web (WASM)
```bash
cd match3
GOOS=js GOARCH=wasm go build -o crystal-cascade.wasm ./cmd

# Запуск с локальным сервером
python3 -m http.server 8080
# Или используйте goexec
go install github.com/shurcooL/goexec@latest
goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
```

## 🎯 Управление

- **Мышь**: клик для выбора фишки, клик на соседнюю для обмена
- **R**: начать заново
- **Подсказки**: появляются автоматически через 5 секунд бездействия

## 🏆 Система очков

- **3 в ряд**: 10 очков за каждую фишку
- **4 в ряд**: 50 очков (бонус)
- **5+ в ряд**: 100 очков (супер бонус)
- **Комбо множитель**: +50% за каждый последовательный матч

## 🎨 Особенности

- Плавные анимации (интерполяция по кадрам)
- Система частиц для визуальных эффектов
- Дрожание фишек при невалидном обмене
- Пульсирующие подсказки
- Прогресс-бар таймера с цветовой индикацией

## 📁 Структура проекта

```
match3/
├── cmd/
│   ├── main.go      # Основной файл игры
│   └── hud.go       # HUD и эффекты частиц
├── sprites/          # Спрайты кристаллов (PNG)
├── web/             # Файлы для Web-версии
├── go.mod           # Go модуль
└── go.sum           # Зависимости
```

## 🛠 Технологии

- **Язык**: Go 1.25
- **Движок**: Ebitengine v2.9.9
- **Графика**: 2D спрайты + векторная отрисовка
- **Платформы**: Windows, Linux, macOS, Android, Web

## 📝 Лицензия

MIT - свободное использование

## 👨‍💻 Автор

Создано в рамках Go365 Challenge - 365 дней изучения Go

---

**Happy Coding! 💎✨**
