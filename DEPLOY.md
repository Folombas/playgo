# 🚀 Purple Lord: Digital Odyssey — Инструкция по запуску

## Быстрый старт

### 1. Запуск сервера (Go)

```bash
cd /home/gofer/godev/projects/playgo/server

# Вариант A: Запуск через go run
go run main.go

# Вариант B: Сборка и запуск
go build -o server main.go
./server
```

**Сервер запустится на:** http://localhost:3000

---

### 2. Запуск клиента

**Вариант A: Python HTTP сервер**
```bash
cd /home/gofer/godev/projects/playgo/client
python3 -m http.server 8080
```

**Вариант B: Go сервер (раздаёт клиент)**
```bash
# Сервер автоматически раздаёт client/ если папка существует
cd /home/gofer/godev/projects/playgo/server
go run main.go
# Открой http://localhost:3000
```

**Вариант C: PHP встроенный сервер**
```bash
cd /home/gofer/godev/projects/playgo/client
php -S localhost:8080
```

**Клиент запустится на:** http://localhost:8080

---

## 🎮 Управление

| Клавиша | Действие |
|---------|----------|
| **A / ←** | Движение влево |
| **D / →** | Движение вправо |
| **W / ↑ / Пробел** | Прыжок |
| **F** | Заклинание (фаербол) |
| **ESC** | Пауза / Меню |

---

## 📊 API Эндпоинты

### Сохранить прогресс
```bash
curl -X POST http://localhost:3000/api/save \
  -H "Content-Type: application/json" \
  -d '{"playerId":"test","totalCrystals":10,"completedLevels":["web_1"]}'
```

### Загрузить прогресс
```bash
curl http://localhost:3000/api/progress?playerId=test
```

### Таблица лидеров
```bash
curl http://localhost:3000/api/leaderboard
```

---

## 🛠️ Требования

- **Go** 1.22+
- **Python 3** (для HTTP сервера) или **PHP**
- **Браузер** с поддержкой JavaScript

---

## 📁 Структура проекта

```
playgo/
├── client/              # Phaser 3 игра
│   ├── index.html
│   ├── assets/
│   └── src/
│       ├── api.js
│       ├── main.js
│       ├── objects/
│       ├── scenes/
│       └── ui/
├── server/              # Go REST API
│   ├── main.go
│   ├── go.mod
│   └── game_data.db     # SQLite (создаётся автоматически)
├── README.md
└── DEPLOY.md
```

---

## 🐛 Отладка

### Включить отладку физики
В `client/src/main.js` изменить:
```javascript
arcade: {
    gravity: { y: 800 },
    debug: true  // Показать хитбоксы
}
```

### Логи сервера
Сервер логирует все запросы:
```
2026/03/15 12:00:00 GET /api/progress 1.234ms
2026/03/15 12:00:01 POST /api/save 2.567ms
```

---

## 🌐 Деплой на сервер (n8n-guru.ru)

### 1. Сборка для Linux
```bash
cd /home/gofer/godev/projects/playgo/server
GOOS=linux GOARCH=amd64 go build -o server main.go
```

### 2. Копирование на сервер
```bash
scp -r playgo user@n8n-guru.ru:/var/www/
```

### 3. Настройка systemd
```bash
sudo nano /etc/systemd/system/purple-lord.service
```

```ini
[Unit]
Description=Purple Lord Game Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/playgo/server
ExecStart=/var/www/playgo/server/server
Restart=always

[Install]
WantedBy=multi-user.target
```

### 4. Запуск
```bash
sudo systemctl daemon-reload
sudo systemctl enable purple-lord
sudo systemctl start purple-lord
sudo systemctl status purple-lord
```

### 5. Nginx конфигурация
```nginx
server {
    listen 80;
    server_name n8n-guru.ru;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

---

## 📝 Заметки

- База данных создаётся автоматически: `server/game_data.db`
- ID игрока сохраняется в localStorage браузера
- Для сброса прогресса очистить localStorage

---

**Go365 Day 75** — 15 марта 2026
