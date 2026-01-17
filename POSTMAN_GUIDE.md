# Postman: проверка сервисов ProfZoom (API + OTP_bot)

Ниже — обновленные сценарии под новую механику: приложение показывает link‑код, пользователь отправляет его боту, бот привязывает Telegram и выдает OTP для входа.

## 1) Подготовка

1. Запустите сервисы и миграции:
   - `docker compose up -d --build`
   - `docker compose up -d migrator`
2. Проверьте, что сервисы поднялись:
   - `GET {{api_base_url}}/health` → `ok`
   - `GET {{otp_base_url}}/health` → `{ "status": "ok" }`
3. Убедитесь, что `.env` заполнены:
   - `Profzom/.env` содержит `OTP_BOT_BASE_URL` и `OTP_BOT_INTERNAL_KEY`
   - `OTP_bot/.env` содержит `API_BASE_URL` и `API_INTERNAL_KEY` (совпадает с `OTP_BOT_INTERNAL_KEY`)

## 2) Импорт коллекции и переменные

1. Импортируйте файл `profzoom.postman_collection.json`.
2. В коллекции задайте переменные:
   - `api_base_url` → `http://localhost:8080`
   - `otp_base_url` → `http://localhost:8081`
   - `otp_internal_key` → ключ `OTP_BOT_INTERNAL_KEY`
   - `user_id` → будет заполнен после `POST /auth/register`
   - `link_code` → будет заполнен после `POST /auth/register`
   - `telegram_id` → chat ID Telegram (для эмуляции можно использовать любое число)
   - `telegram_webhook_secret` → если используете секрет вебхука
   - `otp_code` → код для ручной проверки

## 3) Сценарий “Первый вход (регистрация)”

### Шаг 1: Получить link‑код и user_id
Запрос: `POST /auth/register`  
Ожидания: `{ "user_id": "...", "link_code": "PZ-XXXXXXX" }`

### Шаг 2: Передать link‑код боту
Варианты:
- Реальный Telegram: отправьте `link_code` в чат бота (можно `/start {{link_code}}`).
- Эмуляция: `POST {{otp_base_url}}/telegram/webhook` с `text: "{{link_code}}"` и `chat.id={{telegram_id}}`.

После этого бот привязывает Telegram и (если доступна отправка) присылает OTP.

### Шаг 3: Получить OTP (эмуляция без Telegram)
Если вы тестируете без реального чата, запросите код напрямую:
- `POST /auth/request-code` с `{ "telegram_id": {{telegram_id}} }`  
  Заголовок: `X-Internal-Key: {{otp_internal_key}}`  
Ожидания: `{ "code": "123456", "expires_at": "..." }`

### Шаг 4: Проверить OTP и получить JWT
Запрос: `POST /auth/verify-code`  
Тело: `{ "user_id": "{{user_id}}", "code": "123456" }`  
Ожидания: `{ "token": "...", "is_new_user": true/false }`

## 4) Сценарий “Повторный вход”

1. Пользователь пишет боту `/code` и получает OTP.
2. В приложении вызывает `POST /auth/verify-code` с `{ "user_id": "{{user_id}}", "code": "123456" }`.

Для эмуляции без Telegram используйте `POST /auth/request-code` с `telegram_id`, как в шаге 3 выше.

## 5) Негативные проверки и лимиты

- Частые запросы кода → `rate_limited`.
- Неверный формат `user_id`/`code` → `validation`.
- Просроченный код (`OTP_TTL`) → `unauthorized`.
- Telegram ещё не привязан → `telegram_not_linked`.

## 6) Что ожидать в ответах

Типовые ответы:
- `200 OK` — успешные операции.
- `400` с `validation` — неверные поля.
- `401 unauthorized` — неверный или отсутствующий внутренний ключ.
- `409 telegram_not_linked` — Telegram ещё не привязан.
- `429 rate_limited` — превышен лимит.
- `502 delivery_failed` — ошибка доставки OTP через Telegram.
