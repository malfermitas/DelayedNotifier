# Delayed Notifier

Сервис для планирования и отправки отложенных уведомлений через Email и Telegram.

## Описание

Delayed Notifier - это микросервисное приложение на Go, которое позволяет:

- Создавать отложенные уведомления с указанием даты и времени отправки
- Отправлять уведомления через Email и Telegram
- Отслеживать статус уведомлений
- Отменять запланированные уведомления
- Управлять уведомлениями через Web-интерфейс

## Архитектура

Проект построен по принципу микросервисной архитектуры и состоит из двух основных компонентов:

### 1. API Service (`cmd/app`)

- HTTP API для создания, просмотра и отмены уведомлений
- Web-интерфейс для управления уведомлениями
- Telegram Bot Reader для приема сообщений от пользователей
- Producer для публикации уведомлений в RabbitMQ
- Consumer для обработки результатов отправки

### 2. Worker Service (`cmd/worker`)

- Consumer для получения уведомлений из RabbitMQ
- Отправка уведомлений через Email и Telegram
- Publisher для публикации результатов отправки

## Технологии

- **Go** - основной язык программирования
- **PostgreSQL** - база данных для хранения уведомлений
- **RabbitMQ** - брокер сообщений для асинхронной обработки
- **Docker & Docker Compose** - контейнеризация и оркестрация
- **Gin** - HTTP web-фреймворк
- **Zerolog** - структурированное логирование

## Быстрый старт

### Требования

- Docker
- Docker Compose
- Telegram Bot Token (для отправки через Telegram)

### Запуск

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd DelayedNotifier
```

2. Создайте файл `.env` в корневой директории:
```env
TELEGRAM_TOKEN=your_telegram_bot_token
TELEGRAM_BOT_USERNAME=your_bot_username
```

3. Запустите сервисы:
```bash
docker-compose up -d
```

4. Откройте Web-интерфейс:
```
http://localhost:8080
```

## Конфигурация

### API Service (`config.yaml`)

```yaml
server:
  port: "8080"                    # Порт HTTP сервера
  enable_ui: true                  # Включает Web-интерфейс на маршруте /

database:
  host: "localhost"               # Хост PostgreSQL
  port: "5432"                    # Порт PostgreSQL
  user: "user"                    # Имя пользователя БД
  password: "password"            # Пароль пользователя БД
  name: "delayed_notifier"        # Имя базы данных

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"  # URL RabbitMQ
```

### Worker Service (`config_worker.yaml`)

```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "5432"
  user: "user"
  password: "password"
  name: "delayed_notifier"

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"

email:
  smtp_host: "smtp.gmail.com"     # SMTP сервер
  smtp_port: 587                  # SMTP порт
  username: "your@email.com"      # Email отправителя
  password: "your_app_password"   # Пароль приложения
  from: "your@email.com"          # Имя отправителя
```

### Переменные окружения

| Переменная | Описание | Обязательная |
|------------|----------|--------------|
| `TELEGRAM_TOKEN` | Токен Telegram Bot | Да |
| `TELEGRAM_BOT_USERNAME` | Username бота (без @) | Да |

## API Endpoints

### Создание уведомления
```http
POST /notify
Content-Type: application/json

{
  "message": "Текст уведомления",
  "send_at": "2024-12-31T23:59:59+03:00",
  "channel": "email" | "telegram",
  "recipient": {
    "email": "recipient@example.com",     # для channel=email
    "chat_id": "123456789",               # можно передать напрямую для channel=telegram
    "user_id": "42"                       # или дать user_id, если chat_id уже привязан к пользователю
  }
}
```

Для привязки Telegram пользователь должен написать боту `/start <user_id>`.

**Ответ:**
```json
{
  "id": "notification_id",
  "status": "scheduled"
}
```

### Получение статуса уведомления
```http
GET /notify/{id}
```

**Ответ:**
```json
{
  "id": "uuid",
  "message": "Текст",
  "channel": "email",
  "send_at": "2024-12-31T23:59:59Z",
  "status": "scheduled" | "sent" | "cancelled" | "failed"
}
```

### Отмена уведомления
```http
DELETE /notify/{id}
```

### Web-интерфейс
```http
GET /
```

Маршрут доступен только при `server.enable_ui: true`.

## Структура проекта

```
DelayedNotifier/
├── cmd/
│   ├── app/                    # API сервис
│   │   └── main.go
│   └── worker/                 # Worker сервис
│       └── main.go
├── internal/
│   ├── config/                 # Конфигурация
│   ├── delivery/               # HTTP handlers и middleware
│   ├── model/                  # Модели данных
│   ├── repository/             # Работа с БД
│   ├── service/                # Бизнес-логика
│   ├── sender/                 # Отправка уведомлений
│   └── message_queue/          # RabbitMQ клиенты
├── templates/                  # HTML шаблоны
├── migrations/                 # SQL миграции
├── docker-compose.yaml         # Docker Compose конфигурация
├── Dockerfile                  # Dockerfile для сервисов
└── config.yaml                 # Конфигурация по умолчанию
```

## Разработка

### Локальный запуск без Docker

1. Запустите PostgreSQL и RabbitMQ
2. Примените миграции из `internal/migrations/`
3. Запустите API:
```bash
go run cmd/app/main.go
```
4. Запустите Worker:
```bash
go run cmd/worker/main.go
```

### Сборка

```bash
# Сборка API
go build -o bin/api cmd/app/main.go

# Сборка Worker
go build -o bin/worker cmd/worker/main.go
```
