# 💬 N8N Chat - Real-time Chat на Go

**Современный веб-чат с WebSocket, комнатами и JWT аутентификацией**

---

## 🚀 Быстрый старт

### 1. Запустить сервер

```bash
cd /home/gofer/godev/projects/playgo
go run ./cmd
```

### 2. Открыть в браузере

**http://localhost:8080**

---

## ✨ Особенности

### Для пользователей:
- 💬 **Real-time сообщения** - мгновенная доставка через WebSocket
- 🏠 **Комнаты** - несколько каналов для общения
- 👥 **Онлайн статус** - видно кто в сети
- ⌨️ **Индикатор набора** - видите когда кто-то печатает
- 🔐 **Безопасность** - JWT токены
- 📱 **Адаптивный** - работает на любых устройствах

### Для разработчиков:
- 🚀 **Go 1.25+** - современный Go
- 🔌 **WebSocket** - full-duplex соединение
- 🔑 **JWT** - stateless аутентификация
- 📦 **Модульная архитектура** - чистый код
- 🌐 **CORS** - правильная настройка
- 💾 **In-memory storage** - быстро и просто

---

## 🏗️ Архитектура

```
playgo/
├── cmd/
│   └── main.go              # HTTP сервер + WebSocket
├── internal/
│   ├── auth/
│   │   └── auth.go          # JWT аутентификация
│   ├── chat/
│   │   └── service.go       # Логика чата
│   ├── models/
│   │   └── models.go        # Модели данных
│   └── websocket/
│       └── hub.go           # WebSocket хаб
└── web/
    └── templates/
        └── index.html       # Веб-интерфейс
```

---

## 🎯 Go-концепты в проекте

| Концепт | Где используется |
|---------|-----------------|
| **Горутины** | Обработка клиентов в WebSocket |
| **Каналы** | Broadcast сообщений между клиентами |
| **sync.RWMutex** | Потокобезопасный доступ к данным |
| **net/http** | HTTP сервер |
| **WebSocket** | Real-time соединение |
| **JSON** | Сериализация сообщений |
| **Middleware** | CORS, логирование |
| **JWT** | Аутентификация |

---

## 📊 API Endpoints

### Аутентификация

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/login` | Вход |
| POST | `/register` | Регистрация |

### Чат

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/rooms` | Список комнат |
| GET | `/api/messages?room=general` | Сообщения комнаты |
| GET | `/api/users` | Все пользователи |
| WS | `/ws?room=general` | WebSocket подключение |

---

## 🔌 WebSocket Messages

### Клиент → Сервер

```json
{
  "type": "message",
  "room_id": "general",
  "payload": "Привет всем!"
}
```

```json
{
  "type": "typing",
  "room_id": "general",
  "payload": true
}
```

### Сервер → Клиент

```json
{
  "type": "message",
  "payload": {
    "id": "...",
    "username": "user123",
    "content": "Привет!",
    "timestamp": "2026-03-12T00:00:00Z"
  }
}
```

```json
{
  "type": "online_users",
  "payload": ["user1", "user2", "user3"]
}
```

---

## 🛠️ Переменные окружения

```bash
# Порт сервера (по умолчанию 8080)
PORT=8080

# Секретный ключ для JWT (обязательно измените в production!)
JWT_SECRET=your-super-secret-key-change-this

# Уровень логирования
LOG_LEVEL=info
```

---

## 🚀 Деплой на сервер

### 1. Скомпилировать

```bash
GOOS=linux GOARCH=amd64 go build -o n8n-chat ./cmd
```

### 2. Скопировать на сервер

```bash
scp n8n-chat root@your-server:/opt/n8n-chat/
```

### 3. Создать systemd service

```ini
# /etc/systemd/system/n8n-chat.service
[Unit]
Description=N8N Chat Server
After=network.target

[Service]
Type=simple
User=n8n-chat
WorkingDirectory=/opt/n8n-chat
ExecStart=/opt/n8n-chat/n8n-chat
Restart=always

[Install]
WantedBy=multi-user.target
```

### 4. Запустить

```bash
sudo systemctl daemon-reload
sudo systemctl enable n8n-chat
sudo systemctl start n8n-chat
sudo systemctl status n8n-chat
```

### 5. Настроить Nginx (опционально)

```nginx
server {
    listen 80;
    server_name n8n-guru.ru;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

---

## 📝 Примеры использования

### 1. Создать бота для уведомлений

```javascript
// Отправка уведомлений в Telegram комнату
POST /api/messages
{
  "room": "telegram-bridge",
  "content": "Новая статья в блоге!"
}
```

### 2. Интеграция с сайтом

```html
<!-- Виджет чата поддержки -->
<iframe src="https://n8n-guru.ru/chat?room=support"></iframe>
```

### 3. Команда разработки

```
Комнаты:
- general - общие вопросы
- tech - технические обсуждения
- go-lang - разработка на Go
- news - объявления
```

---

## 🎨 Кастомизация

### Изменить цвета

Откройте `web/templates/index.html` и найдите:

```css
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
```

Замените на свои цвета!

### Добавить комнаты

В `internal/chat/service.go`:

```go
defaultRooms := []struct {
    name        string
    description string
}{
    {"your-room", "Your custom room"},
    // ...
}
```

---

## 🧪 Тестирование

```bash
# Запустить тесты (когда будут добавлены)
go test ./...

# Проверить покрытие
go test -cover ./...
```

---

## 📈 Roadmap

- [ ] SQLite для хранения сообщений
- [ ] Загрузка файлов
- [ ] Эмодзи реакции
- [ ] Поиск по сообщениям
- [ ] Приватные сообщения (DM)
- [ ] Уведомления на email
- [ ] Rate limiting
- [ ] Модерация чата

---

## 📝 Лицензия

MIT

---

**Создано в рамках Go365 челленджа** 🚀

**Демо:** https://n8n-guru.ru

**Автор:** [@Folombas](https://github.com/Folombas)
