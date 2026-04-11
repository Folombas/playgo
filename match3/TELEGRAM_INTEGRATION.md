# 🤖 Telegram Mini App - Match-3 Game

## 📋 Обзор

Интеграция игры "Три в ряд" в Telegram Mini Apps позволяет запускать игру прямо в Telegram без установки дополнительных приложений.

## 🎯 Преимущества

- **Мгновенный доступ** - не нужно скачивать приложение
- **Кроссплатформенность** - работает на любом устройстве с Telegram
- **Встроенная аудитория** - более 800 миллионов пользователей Telegram
- **Простая монетизация** - через Telegram Stars

## 🔧 Техническая реализация

### 1. Web App URL

Игра компилируется в WebAssembly (WASM) и размещается на HTTPS сервере:

```
https://your-domain.com/match3/
```

### 2. Настройка Telegram Bot

```javascript
// Создайте бота через @BotFather
// Отправьте команду /newapp
// Укажите URL вашей WebAssembly игры

// Пример конфигурации:
{
  "short_name": "match3",
  "description": "Три в ряд - классическая головоломка",
  "photo_url": "https://your-domain.com/icon.png",
  "start_param": "play",
  "url": "https://your-domain.com/match3/"
}
```

### 3. Интеграция с Telegram WebApp API

```javascript
// В web/index.html добавьте:
<script src="https://telegram.org/js/telegram-web-app.js"></script>

<script>
// Инициализация
const tg = window.Telegram.WebApp;

// Настройка темы
tg.ready();
tg.expand();

// Получение данных пользователя
const user = tg.initDataUnsafe.user;
console.log('Player:', user.first_name);

// Отправка результатов
function sendScore(score, level) {
  const data = JSON.stringify({
    action: 'score',
    score: score,
    level: level
  });
  
  tg.sendData(data);
}
</script>
```

### 4. Компенсация WASM в Telegram

Telegram WebApp имеет ограничения:
- Максимум 50MB для WASM файлов
- Ограничения на выполнение в фоне
- Нужна оптимизация для мобильных устройств

**Решение:**
```bash
# Компактная WASM сборка
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o match3.wasm ./cmd

# Сжатие
gzip match3.wasm
```

## 📱 Запуск в Telegram

### Шаг 1: Деплой Web версии

```powershell
# Сборка WASM
.\build-web.ps1

# Деплой на GitHub Pages или Vercel
# Например: https://folombas.github.io/playgo/match3/
```

### Шаг 2: Создание Mini App

1. Откройте @BotFather в Telegram
2. Отправьте `/newapp`
3. Следуйте инструкциям
4. Укажите URL игры

### Шаг 3: Тестирование

```
t.me/YourBotName/match3
```

## 🎮 Специальные возможности Telegram

### Deep Linking

```go
// Запуск с параметрами
// t.me/YourBot/match3?startapp=level_5
```

### Haptic Feedback

```javascript
// Вибрация при матчах
if (tg.HapticFeedback) {
  tg.HapticFeedback.impactOccurred('medium');
}
```

### Main Button

```javascript
// Показ основной кнопки
tg.MainButton.setText("Продолжить");
tg.MainButton.show();
tg.MainButton.onClick(() => {
  // Действие
});
```

## 💰 Монетизация

### Telegram Stars

```javascript
// Покупка дополнительных ходов
async function buyMoves(count) {
  await tg.openInvoice({
    currency: 'XTR',
    prices: [
      { label: `${count} moves`, amount: count * 10 }
    ]
  });
}
```

## 📊 Аналитика

```javascript
// Отправка событий
tg.sendData(JSON.stringify({
  event: 'level_complete',
  level: currentLevel,
  score: finalScore,
  moves: totalMoves
}));
```

## 🚀 Roadmap

- [ ] Полная интеграция с Telegram API
- [ ] Таблица лидеров через Telegram
- [ ] Соревнования с друзьями
- [ ] Ежедневные бонусы через Telegram
- [ ] Интеграция с Telegram Premium

## 📝 Пример BotFather конфигурации

```
Bot Name: Match-3 Game
Bot Username: match3_puzzle_bot
Description: Классическая головоломка "Три в ряд" прямо в Telegram!
About: Играйте в Match-3 бесплатно! Собирайте камни, проходите уровни, соревнуйтесь с друзьями.

Commands:
/start - Начать игру
/leaderboard - Таблица лидеров
/daily - Ежедневный бонус
```

## 🔗 Ссылки

- [Telegram WebApps Documentation](https://core.telegram.org/bots/webapps)
- [Telegram BotFather](https://t.me/BotFather)
- [WebAssembly Performance Tips](https://github.com/golang/go/wiki/WebAssembly)

---

**Статус**: 📧 Документация готова
**Дата**: 11 апреля 2026
**Автор**: Go365 Challenge - День 101
