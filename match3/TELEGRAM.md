# Telegram Mini App Configuration
# This file contains the setup instructions for Telegram Mini App

## Telegram Mini App Setup

### 1. Создание бота в Telegram

1. Откройте @BotFather в Telegram
2. Отправьте `/newbot`
3. Следуйте инструкциям для создания бота
4. Сохраните токен бота

### 2. Настройка Mini App

1. В @BotFather отправьте `/newapp`
2. Выберите вашего бота
3. Укажите URL web-приложения (должен быть HTTPS)
4. Укажите короткое название для кнопки

### 3. Деплой web-версии

```bash
# Собрать WASM версию
.\build-web.ps1

# Деплой на хостинг с HTTPS (например, Vercel, Netlify, GitHub Pages)
# Или использовать ngrok для локального тестирования:
ngrok http 8080
```

### 4. Настройка Web App URL

В @BotFather:
```
/mybots -> Выберите бота -> Bot Settings -> Menu Button -> Configure menu button
URL: https://your-domain.com
Text: Играть
```

### 5. Интеграция с Telegram Web App SDK

Для полной интеграции с Telegram (получение пользователя, темы и т.д.),
добавьте в `web/index.html` перед закрывающим `</head>`:

```html
<script src="https://telegram.org/js/telegram-web-app.js"></script>
<script>
  // Инициализация Telegram Web App
  const tg = window.Telegram.WebApp;
  tg.ready();
  
  // Получение данных пользователя
  const user = tg.initDataUnsafe?.user;
  if (user) {
    console.log('User:', user.first_name);
  }
  
  // Адаптация под тему Telegram
  document.body.style.backgroundColor = tg.backgroundColor;
</script>
```

### 6. Команда для запуска

Создайте файл `telegram-deploy.ps1`:

```powershell
# telegram-deploy.ps1
Write-Host "Deploying to Telegram Mini App..." -ForegroundColor Cyan

# 1. Build WASM
.\build-web.ps1

# 2. Деплой (пример для GitHub Pages)
# git add web/
# git commit -m "Update Telegram Mini App"
# git push

# 3. Или деплой на Vercel
# vercel --prod

Write-Host "Don't forget to update the Mini App URL in @BotFather!" -ForegroundColor Yellow
```

## Пример.bot файла

Создайте файл `match3-bot.json` для конфигурации:

```json
{
  "bot_name": "match3_game_bot",
  "app_name": "Match-3 Три в ряд",
  "app_url": "https://your-username.github.io/playgo/match3/web/",
  "description": "Классическая игра Три в ряд на Go + Ebitengine",
  "photo_url": "",
  "start_param": "play"
}
```

## Локальное тестирование

```bash
# Запуск локального сервера
cd web
python -m http.server 8080

# В另一个 терминале запустить ngrok
ngrok http 8080

# Скопировать HTTPS URL из ngrok и указать в @BotFather
```
