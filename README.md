# Order Processing Platform

Учебный микросервисный проект на Go: gRPC сервис заказов, HTTP gateway, события через Kafka, хранение в PostgreSQL и Redis. Проект демонстрирует outbox, идемпотентную обработку событий и базовый observability стек.

## Архитектура

![architecture](docs/order-schema.png)

## Что реализовано

- gRPC API для создания и получения заказа
- HTTP gateway (REST -> gRPC)
- Kafka pipeline:
  - `order-topic` — событие создания заказа
  - `status-topic` — результат оплаты
- Outbox паттерн для надежной публикации событий (orders, payments)
- Идемпотентная обработка Kafka сообщений в payments (`processed_events`)
- Redis хранилище уведомлений с TTL (настраивается через `REDIS_TTL`)
- Метрики Prometheus + Grafana

Protobuf contracts: `https://github.com/ChernykhITMO/order-processing-proto`

## Сервисы

- **orders**
  - gRPC сервис заказов
  - PostgreSQL (orders + order_items)
  - Outbox публикация `OrderCreated`

- **payments**
  - Kafka consumer `order-topic`
  - PostgreSQL (payments + outbox)
  - Публикация `PaymentStatus`

- **notifications**
  - Kafka consumer `status-topic`
  - Redis хранилище уведомлений

- **gateway**
  - HTTP REST API
  - Проксирование в orders по gRPC

- **monitoring**
  - Prometheus
  - Grafana

## Поток событий

1. Gateway принимает REST запрос на создание заказа.
2. Orders сохраняет заказ и пишет `OrderCreated` в outbox.
3. Outbox sender публикует `OrderCreated` в Kafka.
4. Payments читает `OrderCreated`, сохраняет оплату и пишет `PaymentStatus` в outbox.
5. Payments sender публикует `PaymentStatus` в Kafka.
6. Notifications читает `PaymentStatus` и сохраняет уведомление в Redis.

## Быстрый старт (Docker)

1) Скопируйте env-шаблоны:

```bash
cp .env.example .env
cp orders/.env.example orders/.env
cp payments/.env.example payments/.env
cp notifications/.env.example notifications/.env
```

2) Поднимите Postgres и создайте базы:

```bash
docker compose -f docker-compose.yaml up -d postgres
make db-create
```

3) Поднимите весь стек:

```bash
docker compose -f docker-compose.yaml up -d --build
```

4) Миграции:

- В Docker (без локального migrate):

```bash
make docker-migrate-up
```

- Или локально (нужен `migrate`):

```bash
make migrate-up
```

## Эндпоинты и доступы

- API Gateway: `http://localhost:8080`
  - `POST /orders`
  - `GET /orders/{id}`
- Swagger: `http://localhost:8080/swagger/index.html`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`
- Kafka UI: `http://localhost:9020`

## Линтер

```bash
make lint
```

## Тесты

В каждом сервисе отдельный `go.mod`:

```bash
cd orders && go test ./...
cd payments && go test ./...
cd notifications && go test ./...
cd gateway && go test ./...
```

## Примечания

- `ORDERS_GRPC_ADDR` в `orders/.env` — это порт gRPC сервиса. Он должен совпадать с портом, который использует gateway и который проброшен в `docker-compose.yaml`.
- Redis TTL задается через `REDIS_TTL` в `notifications/.env` (например `48h`).
