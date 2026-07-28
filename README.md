# SubscribeREST

Тестовое задание для отклика на вакансию Golang-разрабочтика компании Stackbridge.

## Стэк

- Go v1.26
  - gin v1.12
  - viper v1.21
  - zerolog v1.35
- Swagger
- PostgreSQL 15.17
- Docker + Compose

## Запуск

Для запуска сервера потребуется Docker Compose:

```sh
docker compose up --build
```

## Эндпоинты

`/api/v1`:

- GET `/subscriptions` — список всех подписок с пагинацией,
- POST `/subscriptions` — создание новой подписки,
- GET `/subscriptions/{id}` — получение подписки по ID,
- PUT `/subscriptions/{id}` — изменение подписки по ID,
- DELETE `/subscriptions/{id}` — удаление подписки по ID,
- GET `/subscriptions/stats` — подсчёт суммарной стоимости подписок пользователя за период с фильтрацией по названию сервиса.
