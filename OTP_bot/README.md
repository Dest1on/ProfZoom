# ProfZoom OTP Bot

OTP_bot - это сервис доставки и привязки Telegram. Он НЕ ДОЛЖЕН генерировать, проверять или хранить OTP коды. Он только отправляет предоставленные OTP коды в чаты Telegram и управляет привязкой аккаунтов Telegram.

## Обязанности

- Только доставка OTP через Telegram.
- Процесс привязки Telegram (регистрация токена, привязка через /start, проверки статуса).
- Один HTTP сервер публикует все эндпоинты на одном base URL: `/telegram/webhook`, `/telegram/link-token`, `/telegram/status`, `/otp/send`, `/health`.

## Переменные окружения

Требуются:

```
TELEGRAM_BOT_TOKEN=replace_me
OTP_BOT_INTERNAL_KEY=replace_me
```

Опционально (показаны значения по умолчанию):

```
DATABASE_URL=postgres://user:pass@localhost:5432/profzoom?sslmode=disable
DB_DRIVER=pgx
TELEGRAM_WEBHOOK_SECRET=replace_me
TELEGRAM_LINK_TTL=10m
LINK_TOKEN_RATE_LIMIT_PER_MIN=5
LINK_TOKEN_RATE_LIMIT_IP_PER_MIN=5
LINK_TOKEN_RATE_LIMIT_BOT_PER_MIN=5
PORT=8080
LOG_LEVEL=info
TELEGRAM_TIMEOUT=5s
OTP_RATE_LIMIT_PER_MIN=2
OTP_RATE_LIMIT_IP_PER_MIN=2
OTP_RATE_LIMIT_BOT_PER_MIN=60
```

Лимиты per-IP/per-bot по умолчанию используют их текущие значения per-minute, если переменные не заданы.
`TELEGRAM_LINK_TTL` должен быть между 5m и 10m.
Если `DATABASE_URL` не задан, сервис использует in-memory хранилища.

## Миграции

Запустите SQL из каталога `migrations/` в вашей базе Postgres.
Таблицы, используемые этим сервисом: `telegram_links`, `telegram_link_tokens`.
`telegram_links` принадлежит этому сервису; если основной бэкенд зеркалит ее, синхронизируйте схемы.

## HTTP эндпоинты

Спецификация OpenAPI доступна в `openapi.yaml`.

### POST /otp/send

Эндпоинт только для доставки OTP.

Headers:

```
X-Internal-Key: ${OTP_BOT_INTERNAL_KEY}
```

Body:

```
{ "phone": "+15551234567", "code": "834291" }
```

Ограничения:

- `phone` должен быть привязан.
- `code` не может быть пустым.
- Сообщение в Telegram отправляется как: `ProfZoom login code: <code>`.

Responses:

- `200` `{ "sent": true }`
- `400` `{ "error": "invalid_payload" }` или `{ "error": "phone_not_linked" }`
- `401` `{ "error": "unauthorized" }`
- `429` `{ "error": "rate_limited" }`
- `500` `{ "error": "telegram_failed" }`

### POST /telegram/link-token

Регистрация одноразового токена привязки Telegram.

Headers:

```
X-Internal-Key: ${OTP_BOT_INTERNAL_KEY}
```

Body:

```
{ "phone": "+15551234567", "token": "..." }
```

Response:

```
{ "success": true }
```

### GET /telegram/status

Возвращает статус привязки Telegram для номера телефона.

Headers:

```
X-Internal-Key: ${OTP_BOT_INTERNAL_KEY}
```

Query:

```
/telegram/status?phone=+15551234567
```

Body (optional):

```
{ "phone": "+15551234567" }
```

Responses:

Привязан:

```
{ "linked": true }
```

Не привязан:

```
{ "linked": false }
```

### POST /telegram/webhook

Эндпоинт Telegram webhook.

Headers:

```
X-Telegram-Bot-Api-Secret-Token: ${TELEGRAM_WEBHOOK_SECRET}
```

Поддерживает `/start <token>`, `/help`, `/status` и отправку контакта в приватных чатах.

### GET /health

Эндпоинт проверки здоровья.

Response:

```
{ "status": "ok" }
```

## Процесс привязки

1. Основной бэкенд вызывает `GET /telegram/status?phone=...`.
2. Если не привязан, основной бэкенд вызывает `POST /telegram/link-token` с `{ phone, token }`.
3. Приложение формирует Telegram deep link: `/start <token>` и отправляет его пользователю.
4. Пользователь открывает ссылку в Telegram; OTP_bot обрабатывает `/telegram/webhook` и связывает чат с телефоном.
5. Основной бэкенд снова вызывает `GET /telegram/status?phone=...`, затем `POST /otp/send`.

Пользователи также могут привязаться, отправив свой контакт боту в приватном чате.

## Запуск локально

```
go run ./cmd/server
```

## CI

```
go test ./...
```
