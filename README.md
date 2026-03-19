# Barber Backend — Go API

Go бэкенд для сайта записи к парикмахеру. Echo + GORM + PostgreSQL.

## Структура

```
barber-backend/
├── cmd/server/main.go              # Точка входа
├── internal/
│   ├── model/appointment.go        # ORM модель
│   ├── repository/appointment.go   # Слой данных
│   ├── service/appointment.go      # Бизнес-логика
│   └── handler/appointment.go      # HTTP хэндлеры
├── .env.example                    # Пример переменных
├── Makefile
├── Dockerfile
└── go.mod
```

## API эндпоинты

| Метод  | URL                          | Описание                    |
|--------|------------------------------|-----------------------------|
| POST   | /api/appointments            | Создать запись              |
| GET    | /api/appointments?date=...   | Записи за день              |
| GET    | /api/appointments/range?start=...&end=... | Записи за период |
| GET    | /api/appointments/slots?date=... | Занятые слоты за дату    |
| DELETE | /api/appointments/:id        | Удалить запись              |
| GET    | /health                      | Health check                |

## Быстрый старт

### 1. Установи PostgreSQL (если нет)

macOS:
```bash
brew install postgresql@16
brew services start postgresql@16
```

### 2. Создай базу данных

```bash
createdb barber
```

### 3. Настрой окружение

```bash
cp .env.example .env
```

Отредактируй `.env` если нужно (по умолчанию localhost:5432, user=postgres).

### 4. Запусти

```bash
go mod tidy
make run
```

Сервер стартует на `http://localhost:8080`.

### 5. Настрой фронтенд

В `.env.local` фронтенда замени Supabase на Go API:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_MASTER_PIN=1234
```

Также замени файлы:
- `src/lib/api.js` — новый API клиент (вместо supabase.js)
- `src/app/page.js` — обновлённая клиентская страница
- `src/app/master/page.js` — обновлённая панель мастера

## Docker

```bash
docker build -t barber-backend .
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/barber \
  barber-backend
```
