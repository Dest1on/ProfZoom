# ProfZoom API

Основной сервис — единый источник истины для авторизации. Он генерирует OTP, хранит только хэши, проверяет OTP, принимает решение об авторизации и выпускает JWT/refresh токены. Доставкой OTP занимается OTP_bot.

## Процесс авторизации (Telegram‑first)

1. Приложение вызывает `POST /auth/register` и получает:
   - `user_id` — идентификатор пользователя (сохраните в приложении)
   - `link_code` — код для привязки Telegram (показывается пользователю)
2. Пользователь отправляет `link_code` боту. Бот привязывает чат к `user_id` и запрашивает OTP.
3. Пользователь получает OTP в Telegram и вводит его в приложении.
4. Приложение вызывает `POST /auth/verify-code` с `{ "user_id": "<uuid>", "code": "123456" }`.

Повторный вход: пользователь пишет боту `/code`, получает OTP и подтверждает его через `/auth/verify-code`.

## Внутренние эндпоинты для бота

- `POST /auth/request-code` с `{ "telegram_id": 123456789 }`
  - Заголовок авторизации: `X-Internal-Key: $OTP_BOT_INTERNAL_KEY`
  - Ответ: `{ "code": "123456", "expires_at": "..." }`
- `POST /auth/verify-code` с `{ "telegram_id": 123456789, "code": "123456" }`

## Интеграция с OTP_bot

- `POST /telegram/link-token` с `{ "user_id": "<uuid>", "token": "<link_code>" }`
  - Заголовок авторизации: `X-Internal-Key: $OTP_BOT_INTERNAL_KEY`
  - Регистрирует link‑код, который пользователь отправляет боту.
- `GET /telegram/status?user_id=<uuid>` (опционально)
  - Заголовок авторизации: `X-Internal-Key: $OTP_BOT_INTERNAL_KEY`

## Токены и сессии

- Access JWT включает `sub` (user_id), `roles`, `exp`, `iat`.
- Refresh токены хранятся в виде хэшей, ротируются при обновлении и отзываются при выходе.

## Роли

- `student`, `company`
- Проверка ролей выполняется middleware.
- Роль выбирается позже через `PATCH /users/role` с `{ "role": "student" | "company" }`.

## Переменные окружения

Требуются:
- `DATABASE_URL`
- `JWT_SECRET`
- `OTP_BOT_BASE_URL`
- `OTP_BOT_INTERNAL_KEY`

Опционально:
- `HTTP_PORT` (по умолчанию `8080`)
- `ACCESS_TOKEN_TTL` (по умолчанию `15m`)
- `REFRESH_TOKEN_TTL` (по умолчанию `720h`)
- `OTP_TTL` (по умолчанию `2m`)
- `DB_MAX_OPEN_CONNS` (по умолчанию `25`)
- `DB_MAX_IDLE_CONNS` (по умолчанию `10`)
- `DB_CONN_MAX_IDLE` (по умолчанию `5m`)
- `DB_CONN_MAX_LIFE` (по умолчанию `30m`)
- `REQUEST_TIMEOUT` (по умолчанию `10s`)
