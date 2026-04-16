# DelayedNotifier

DelayedNotifier — сервис для создания и управления отложенными уведомлениями через различные каналы связи (Email, Telegram).  
Проект поддерживает гибкое планирование доставки, отмену запланированных уведомлений и отслеживание их статуса в реальном времени.

## Возможности

- Создание отложенных уведомлений (Email, Telegram).
- Просмотр очереди активных и завершенных уведомлений.
- Проверка статуса уведомления по ID.
- Отмена запланированного уведомления до момента его отправки.
- Асинхронная обработка через RabbitMQ.
- Кэширование данных для быстрого доступа через Redis.
- Современный веб-интерфейс (React + Vite).

## Технологический стек

- Язык программирования: Go 1.25.
- Web-фреймворк: Ginext.
- База данных: PostgreSQL 16.
- Брокер сообщений: RabbitMQ (с поддержкой DLX).
- Кэш: Redis.
- Миграции: Migrate.
- Логирование: Zerolog (через wbf).
- Контейнеризация: Docker, Docker Compose.

## Структура проекта

```text
.
├── cmd                 # Точка входа в приложение
├── configs             # Конфигурационные файлы
├── internal            # Внутренняя логика проекта
│   ├── app            # Инициализация и сборка зависимостей
│   ├── config         # Загрузка конфигурации
│   ├── domain         # Доменные модели
│   ├── infra          # Инфраструктурные клиенты (RabbitMQ, Redis, Postgres)
│   ├── logger         # Адаптер логирования
│   ├── repository     # Слой работы с БД
│   ├── service        # Бизнес-логика
│   └── transport      # HTTP роутинг и хендлеры
├── migrations          # SQL-миграции
└── web                 # Фронтенд-приложение (React)
    ├── node_modules
    └── src
```

## Требования

- Docker.
- Docker Compose.
- Make.
- Go 1.25.
- Node.js и npm (для локальной разработки фронтенда).

## Установка и запуск

### Подготовка окружения

Перед запуском необходимо создать файл `.env` на основе примера:

```bash
cp .env.example .env
```
Заполните необходимые переменные (токены Telegram, параметры почты) в созданном файле.

### Быстрый старт

```bash
make up
```
Фронтенд-приложение доступно по адресу http://localhost:3000 (Vite dev server, работающий внутри контейнера).

## Миграции

Миграции применяются автоматически при запуске контейнеров, но их можно вызвать вручную:

### Применить миграции

```bash
make migrate-up
```

### Откатить одну миграцию

```bash
make migrate-down
```

## Тестирование

### Запуск тестов

```bash
make test
```

### Запуск тестов с покрытием

```bash
go test ./... -cover
```

## Форматирование и линтинг

### Форматирование Go-кода

```bash
make fmt
```

### Проверка линтером

```bash
make lint
```

## HTTP API

### Маршруты

```go
r.POST("/create", handlers.Notifications.Create)
r.GET("/all", handlers.Notifications.GetAll)
r.GET("/status/:id", handlers.Notifications.GetStatus)
r.DELETE("/cancel/:id", handlers.Notifications.Cancel)
```

### Эндпоинты

| Method | Route          | Description                     |
|--------|----------------|---------------------------------|
| POST   | `/create`      | Создать отложенное уведомление  |
| GET    | `/all`         | Получить список всех уведомлений|
| GET    | `/status/:id`  | Проверить статус по ID          |
| DELETE | `/cancel/:id`  | Отменить уведомление            |

## Примеры запросов

### Создать уведомление

```bash
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{
    "destination": "user@example.com",
    "channel": "email",
    "message": "Hello! This is a delayed message.",
    "data_sent_at": "2026-04-17T10:00:00Z"
  }'
```

### Отменить уведомление

```bash
curl -X DELETE http://localhost:8080/cancel/8d14d543-297b-4cd2-9c0d-1396acd146a5
```

## Конфигурация

Конфигурация находится в файле `configs/config.yml`. Часть параметров (секреты) задаются через `.env` файл. 

## Полезные команды

```bash
make help      # Список всех команд
make up        # Запуск проекта
make down      # Остановка проекта (с удалением volume используйте make down -v)
make logs      # Просмотр логов
make restart   # Перезапуск сервисов
```
