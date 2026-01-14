# ProfZoom API

Основной сервис является единым источником истины для авторизации. Он генерирует OTP, хранит только хэши, проверяет OTP, принимает решение об авторизации и выпускает JWT/refresh токены. Доставкой OTP занимается исключительно OTP_bot.

## Процесс авторизации

- `POST /auth/send-code` с `{ "phone": "+79991234567" }`
  - Генерирует OTP и сохраняет хэш с TTL и оставшимся числом попыток.
  - Проверяет привязку Telegram через OTP_bot.
  - Если Telegram не привязан, регистрирует link-token и возвращает `need_link`.
  - Отправляет OTP через OTP_bot.
  - Ответ, когда нужна привязка: `{ "success": false, "need_link": true, "telegram_token": "<token>", "telegram_link": "https://t.me/ProfZoomOtpBot?start=<token>" }`
- `POST /auth/verify-code` с `{ "phone": "+79991234567", "code": "123456" }`
  - Проверяет хэш OTP и выдает access + refresh токены.
  - Ответ: `{ "token": "<access_jwt>", "refresh_token": "<refresh>", "is_new_user": true }`

## Интеграция с OTP_bot

- `GET /telegram/status?phone=+79991234567` с заголовком `X-Internal-Key: $OTP_BOT_INTERNAL_KEY`
  - 200 linked: `{ "linked": true }`
  - 200 not linked: `{ "linked": false }`
- `POST /telegram/link-token` с `{ "phone": "+79991234567", "token": "<token>" }`
  - Заголовок авторизации: `X-Internal-Key: $OTP_BOT_INTERNAL_KEY`
- `POST /otp/send` с `{ "phone": "+79991234567", "code": "123456" }`
  - Заголовок авторизации: `X-Internal-Key: $OTP_BOT_INTERNAL_KEY`
- Сервис не хранит Telegram `chat_id`; статус запрашивается на каждый запрос.

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
- `OTP_BOT_TELEGRAM_USERNAME` (username бота для генерации deep link)
