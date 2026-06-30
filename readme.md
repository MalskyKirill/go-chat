# Real-time Chat на Go

Учебный backend-проект real-time чата на Go.

Проект реализует регистрацию пользователей, авторизацию через JWT, личные и групповые чаты, историю сообщений, WebSocket-подключения, отправку сообщений в реальном времени и online-статус пользователей.

## Возможности

На текущий момент реализовано:

* регистрация пользователей;
* логин пользователей;
* JWT-авторизация;
* защищённые REST endpoints;
* создание личных чатов;
* создание групповых чатов;
* получение списка чатов пользователя;
* отправка сообщений через REST;
* получение истории сообщений;
* WebSocket-подключение;
* отправка сообщений через WebSocket;
* сохранение WebSocket-сообщений в PostgreSQL;
* рассылка новых сообщений всем online-участникам чата;
* online/offline-статус пользователей;
* получение online-пользователей из общих чатов;
* получение online-участников конкретного чата.

## Стек

* Go
* PostgreSQL
* Docker Compose
* JWT
* bcrypt
* WebSocket
* gorilla/websocket
* pgx
* godotenv

## Структура проекта

```text
real-time-chat/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── jwt.go
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   └── postgres.go
│   ├── dto/
│   │   ├── auth.go
│   │   ├── chat.go
│   │   ├── message.go
│   │   └── online.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── chat_handler.go
│   │   ├── health_handler.go
│   │   ├── message_handler.go
│   │   ├── online_handler.go
│   │   ├── response.go
│   │   └── ws_handler.go
│   ├── middleware/
│   │   └── auth_middleware.go
│   ├── models/
│   │   ├── chat.go
│   │   ├── message.go
│   │   └── user.go
│   ├── repositories/
│   │   ├── chat_repository.go
│   │   ├── errors.go
│   │   ├── message_repository.go
│   │   └── user_repository.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── chat_service.go
│   │   └── message_service.go
│   └── websocket/
│       ├── client.go
│       ├── hub.go
│       └── message.go
├── migrations/
│   ├── 000_init.sql
│   ├── 001_create_users.sql
│   ├── 002_create_chats.sql
│   └── 003_create_messages.sql
├── docker-compose.yml
├── Dockerfile
├── .env
├── .env.example
├── .gitignore
├── go.mod
└── go.sum
```

## Переменные окружения

Пример `.env`:

```env
HTTP_PORT=8080

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=chat_user
POSTGRES_PASSWORD=chat_password
POSTGRES_DB=chat_db

JWT_SECRET=super-secret-dev-key
JWT_TTL_HOURS=24
```

Для Docker Compose приложение использует сервис PostgreSQL по имени `postgres`.

## Запуск проекта

Установить зависимости:

```bash
go mod tidy
```

Запустить проект через Docker Compose:

```bash
docker compose up --build
```

Если нужно пересоздать базу данных с нуля:

```bash
docker compose down -v
docker compose up --build
```

Проверить, что сервер работает:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{
  "database": "ok",
  "status": "ok"
}
```

## REST API

### Health check

```http
GET /health
```

Пример:

```bash
curl http://localhost:8080/health
```

Ответ:

```json
{
  "database": "ok",
  "status": "ok"
}
```

---

## Авторизация

### Регистрация

```http
POST /api/auth/register
```

Тело запроса:

```json
{
  "username": "kirill",
  "email": "kirill@example.com",
  "password": "123456"
}
```

Пример:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "kirill",
    "email": "kirill@example.com",
    "password": "123456"
  }'
```

Ответ:

```json
{
  "token": "JWT_TOKEN",
  "user": {
    "id": 1,
    "username": "kirill",
    "email": "kirill@example.com",
    "created_at": "2026-06-30T12:00:00Z"
  }
}
```

### Логин

```http
POST /api/auth/login
```

Тело запроса:

```json
{
  "email": "kirill@example.com",
  "password": "123456"
}
```

Пример:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "kirill@example.com",
    "password": "123456"
  }'
```

Ответ:

```json
{
  "token": "JWT_TOKEN",
  "user": {
    "id": 1,
    "username": "kirill",
    "email": "kirill@example.com",
    "created_at": "2026-06-30T12:00:00Z"
  }
}
```

### Получить текущего пользователя

```http
GET /api/me
```

Пример:

```bash
curl http://localhost:8080/api/me \
  -H "Authorization: Bearer JWT_TOKEN"
```

Ответ:

```json
{
  "id": 1,
  "username": "kirill",
  "email": "kirill@example.com",
  "created_at": "2026-06-30T12:00:00Z"
}
```

---

## Чаты

### Создать личный чат

```http
POST /api/chats/private
```

Тело запроса:

```json
{
  "user_id": 2
}
```

Пример:

```bash
curl -X POST http://localhost:8080/api/chats/private \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer JWT_TOKEN" \
  -d '{
    "user_id": 2
  }'
```

Ответ:

```json
{
  "id": 1,
  "type": "private",
  "title": "",
  "members": [
    {
      "id": 1,
      "username": "kirill"
    },
    {
      "id": 2,
      "username": "ivan"
    }
  ],
  "created_at": "2026-06-30T12:00:00Z"
}
```

Если личный чат между этими пользователями уже существует, сервер вернёт существующий чат.

### Создать групповой чат

```http
POST /api/chats/group
```

Тело запроса:

```json
{
  "title": "Go Developers",
  "member_ids": [2, 3]
}
```

Пример:

```bash
curl -X POST http://localhost:8080/api/chats/group \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer JWT_TOKEN" \
  -d '{
    "title": "Go Developers",
    "member_ids": [2, 3]
  }'
```

Ответ:

```json
{
  "id": 2,
  "type": "group",
  "title": "Go Developers",
  "members": [
    {
      "id": 1,
      "username": "kirill"
    },
    {
      "id": 2,
      "username": "ivan"
    },
    {
      "id": 3,
      "username": "anna"
    }
  ],
  "created_at": "2026-06-30T12:00:00Z"
}
```

Создатель группового чата автоматически добавляется в участники.

### Получить список своих чатов

```http
GET /api/chats
```

Пример:

```bash
curl http://localhost:8080/api/chats \
  -H "Authorization: Bearer JWT_TOKEN"
```

Ответ:

```json
[
  {
    "id": 1,
    "type": "private",
    "title": "",
    "members": [
      {
        "id": 1,
        "username": "kirill"
      },
      {
        "id": 2,
        "username": "ivan"
      }
    ],
    "created_at": "2026-06-30T12:00:00Z"
  }
]
```

---

## Сообщения

### Отправить сообщение через REST

```http
POST /api/chats/{chatID}/messages
```

Тело запроса:

```json
{
  "content": "Привет!"
}
```

Пример:

```bash
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer JWT_TOKEN" \
  -d '{
    "content": "Привет!"
  }'
```

Ответ:

```json
{
  "id": 1,
  "chat_id": 1,
  "sender_id": 1,
  "sender_username": "kirill",
  "content": "Привет!",
  "created_at": "2026-06-30T12:00:00Z"
}
```

### Получить историю сообщений

```http
GET /api/chats/{chatID}/messages
```

Пример:

```bash
curl http://localhost:8080/api/chats/1/messages \
  -H "Authorization: Bearer JWT_TOKEN"
```

Ответ:

```json
[
  {
    "id": 1,
    "chat_id": 1,
    "sender_id": 1,
    "sender_username": "kirill",
    "content": "Привет!",
    "created_at": "2026-06-30T12:00:00Z"
  }
]
```

Можно использовать пагинацию:

```bash
curl "http://localhost:8080/api/chats/1/messages?limit=20&offset=0" \
  -H "Authorization: Bearer JWT_TOKEN"
```

Ограничения:

```text
limit по умолчанию = 50
максимальный limit = 100
offset по умолчанию = 0
```

---

## Online-статус

### Получить online-пользователей из общих чатов

```http
GET /api/users/online
```

Пример:

```bash
curl http://localhost:8080/api/users/online \
  -H "Authorization: Bearer JWT_TOKEN"
```

Ответ:

```json
[
  {
    "id": 2,
    "username": "ivan"
  }
]
```

Этот endpoint возвращает online-пользователей, у которых есть общий чат с текущим пользователем.

### Получить online-участников конкретного чата

```http
GET /api/chats/{chatID}/online
```

Пример:

```bash
curl http://localhost:8080/api/chats/1/online \
  -H "Authorization: Bearer JWT_TOKEN"
```

Ответ:

```json
[
  {
    "id": 1,
    "username": "kirill"
  },
  {
    "id": 2,
    "username": "ivan"
  }
]
```

---

## WebSocket

WebSocket endpoint:

```text
GET /ws?token=JWT_TOKEN
```

Пример подключения через `wscat`:

```bash
wscat -c "ws://localhost:8080/ws?token=JWT_TOKEN"
```

После подключения сервер отправит:

```json
{
  "type": "connection.ready",
  "data": {
    "user_id": 1
  }
}
```

Когда пользователь становится online, пользователи из общих чатов получают событие:

```json
{
  "type": "user.online",
  "data": {
    "user_id": 2
  }
}
```

Когда пользователь становится offline:

```json
{
  "type": "user.offline",
  "data": {
    "user_id": 2
  }
}
```

Пользователь считается offline только после закрытия всех его WebSocket-подключений.

---

## Отправка сообщений через WebSocket

Клиент отправляет:

```json
{
  "type": "message.send",
  "chat_id": 1,
  "content": "Привет через WebSocket!"
}
```

Сервер сохраняет сообщение в PostgreSQL и рассылает всем online-участникам чата:

```json
{
  "type": "message.new",
  "data": {
    "id": 1,
    "chat_id": 1,
    "sender_id": 1,
    "sender_username": "kirill",
    "content": "Привет через WebSocket!",
    "created_at": "2026-06-30T12:00:00Z"
  }
}
```

Сообщение, отправленное через WebSocket, также доступно в истории сообщений через REST:

```http
GET /api/chats/{chatID}/messages
```

---

## WebSocket ошибки

Если клиент отправил невалидный JSON:

```json
{
  "type": "error",
  "data": {
    "code": "invalid_json",
    "message": "invalid json message"
  }
}
```

Если тип события неизвестен:

```json
{
  "type": "error",
  "data": {
    "code": "unknown_event",
    "message": "unknown event type"
  }
}
```

Если сообщение пустое:

```json
{
  "type": "error",
  "data": {
    "code": "message_empty",
    "message": "message content is empty"
  }
}
```

Если пользователь не состоит в чате:

```json
{
  "type": "error",
  "data": {
    "code": "forbidden",
    "message": "you are not a member of this chat"
  }
}
```

Если чат не найден:

```json
{
  "type": "error",
  "data": {
    "code": "chat_not_found",
    "message": "chat not found"
  }
}
```

---

## Быстрая проверка проекта

### 1. Запустить проект

```bash
docker compose up --build
```

### 2. Зарегистрировать пользователя Kirill

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "kirill",
    "email": "kirill@example.com",
    "password": "123456"
  }'
```

### 3. Зарегистрировать пользователя Ivan

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ivan",
    "email": "ivan@example.com",
    "password": "123456"
  }'
```

### 4. Создать личный чат

```bash
curl -X POST http://localhost:8080/api/chats/private \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN_KIRILL" \
  -d '{
    "user_id": 2
  }'
```

### 5. Открыть WebSocket для Kirill

```bash
wscat -c "ws://localhost:8080/ws?token=TOKEN_KIRILL"
```

### 6. Открыть WebSocket для Ivan

```bash
wscat -c "ws://localhost:8080/ws?token=TOKEN_IVAN"
```

### 7. Отправить сообщение через WebSocket

```json
{
  "type": "message.send",
  "chat_id": 1,
  "content": "Привет, Иван!"
}
```

### 8. Проверить историю сообщений

```bash
curl http://localhost:8080/api/chats/1/messages \
  -H "Authorization: Bearer TOKEN_KIRILL"
```

---

## База данных

### users

Хранит пользователей.

Основные поля:

```text
id
username
email
password_hash
created_at
```

### chats

Хранит чаты.

Основные поля:

```text
id
type
title
created_at
```

`type` может быть:

```text
private
group
```

### chat_members

Хранит участников чатов.

Основные поля:

```text
chat_id
user_id
created_at
```

### messages

Хранит сообщения.

Основные поля:

```text
id
chat_id
sender_id
content
created_at
```

---

## Текущий статус проекта

Готово:

```text
Этап 1 — каркас проекта
Этап 2 — регистрация и авторизация
Этап 3 — чаты
Этап 4 — сообщения через REST
Этап 5 — WebSocket-подключение
Этап 6 — сообщения через WebSocket
Этап 7 — online-статус
```

## Что можно улучшить позже

* добавить уведомления;
* добавить индикатор "печатает";
* добавить read receipts;
* добавить редактирование сообщений;
* добавить удаление сообщений;
* добавить загрузку файлов;
* добавить Redis Pub/Sub для масштабирования WebSocket;
* добавить rate limiting;
* добавить frontend на HTML/CSS/JS;
* добавить тесты.
