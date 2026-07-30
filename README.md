# SubscribeREST

Тестовое задание для отклика на вакансию Golang-разрабочтика компании Stackbridge.

## Стэк

- Go v1.26
  - gin v1.12
  - viper v1.21
  - zerolog v1.35
  - golang-migrate v4
  - testify (тесты)
- Swagger
- PostgreSQL 15.17
- Docker Compose

## Запуск

Для запуска сервера потребуется Docker Compose:

```sh
docker compose up --build
```

После запуска Swagger UI доступен по адресу http://localhost:8080/swagger/index.html

## Конфигурация

Настройки порта сервера и подключения к базе данных задаются через переменные окружения в [`compose.yaml`](compose.yaml):

| Переменная | Описание | Значение по умолчанию |
|---|---|---|
| `PORT` | Порт сервера | `8080` |
| `DB_HOST` | Хост БД | `database` |
| `DB_PORT` | Порт БД | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `postgres` |
| `DB_NAME` | Имя БД | `subscriberest_dev` |
| `DB_SSLMODE` | SSL mode | `disable` |

## Эндпоинты

`/api/v1`:

- GET `/subscriptions` — список всех подписок с пагинацией,
- POST `/subscriptions` — создание новой подписки,
- GET `/subscriptions/{id}` — получение подписки по ID,
- PUT `/subscriptions/{id}` — изменение подписки по ID,
- DELETE `/subscriptions/{id}` — удаление подписки по ID,
- GET `/subscriptions/stats` — подсчёт суммарной стоимости подписок пользователя за период с фильтрацией по названию сервиса.

## Разработка

### Makefile

| Команда | Описание |
|---|---|
| `make generate` | Регенерировать Swagger-спецификацию |
| `make build` | Сгенерировать спецификацию и собрать бинарник |
| `make test` | Запустить все тесты |

### Миграции БД

Применяются автоматически при старте сервера (golang-migrate, `migrations/`).

Для ручного управления ([golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)):

```sh
make migrate-up
make migrate-down
```

### Swagger-документация

OpenAPI-спецификация генерируется автоматически с помощью swaggo/swag.

Регенерация:

```sh
make generate
```

Сгенерированные файлы (`docs/`) находятся в репозитории и импортируются в `cmd/server/main.go`.

### Тестирование

```sh
make test
```
