# Репозиторий с домашними заданиями по курсу МФТИ язык Go.
## HW1.go - первая домашняя работа.
## Добавлены модули в рамках второй домашней работы:
- Gateway - модуль шлюза согласно ДЗ2
- Ledger - модуль логики согласно ДЗ2
## Домашнее задание 7 (ДЗ7)
### Тема: Интеграция с PostgreSQL и Redis
###Что реализовано:
- PostgreSQL: хранение бюджетов и транзакций.
- Redis: кэширование отчётов и списков бюджетов.
- Goose: миграции базы данных.
- Connection pooling: настроен пул соединений для устойчивости к нагрузке.
### Установка и запуск
1. Установка зависимостей
Установите утилиту Goose:
- go install github.com/pressly/goose/v3/cmd/goose@latest
Запустите PostgreSQL и Redis (например, через Docker):
- docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=cashapp postgres:15-alpine
- docker run -d --name redis -p 6379:6379 redis:7-alpine
2. Применение миграций
Перейдите в папку миграций:
- cd ledger/migrations
Примените миграции:
- goose -dir . postgres "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable" up
Откат миграций (при необходимости):
- goose -dir . postgres "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable" down
3. Переменные окружения
Для PostgreSQL (приоритетный вариант):
- export DATABASE_URL="postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
Для Redis:
- export REDIS_ADDR=localhost:6379
- export REDIS_DB=0
- export REDIS_PASSWORD=""
4. Запуск сервисов
Запуск Ledger (из папки gateway):
- cd gateway
- go mod tidy
- go run main.go
### API Endpoints
Бюджеты
POST /api/budgets — Создать или обновить бюджет
- curl -X POST http://localhost:8080/api/budgets -H "Content-Type: application/json" -d '{"category":"Food","limit":5000,"period":"month"}'
GET /api/budgets — Получить список бюджетов (с кэшированием)
- curl http://localhost:8080/api/budgets
Транзакции
POST /api/transactions — Создать транзакцию
- curl -X POST http://localhost:8080/api/transactions -H "Content-Type: application/json" -d '{"amount":1000,"category":"Food","description":"Lunch","date":"2026-10-10"}'
GET /api/transactions — Получить список транзакций
- curl http://localhost:8080/api/transactions
Отчёты (с кэшированием)
GET /api/reports/summary — Получить сводный отчёт
Первый запрос (cache miss, расчёт из БД):
- curl "http://localhost:8080/api/reports/summary?from=2026-10-01&to=2026-10-30"
Второй запрос (cache hit, ответ из Redis):
- curl "http://localhost:8080/api/reports/summary?from=2026-10-01&to=2026-10-30"
### Проверка работы
Данные сохраняются в PostgreSQL и переживают перезапуск сервиса.
Отчёты кэшируются в Redis на 30 секунд.
При превышении бюджета возвращается ошибка 409 Conflict.
Gateway обращается к тем же HTTP-эндпоинтам, но теперь Ledger хранит данные в PostgreSQL, а не в памяти.
### Тестирование
Запуск тестов:
- export DATABASE_URL="postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
- go test ledger/... gateway/... -v -cover

