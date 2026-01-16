# Order Processing Platform
Система обработки заказов с gRPC-взаимодействием, событийной коммуникацией через Kafka, PostgreSQL и полноценным мониторингом в Prometheus + Grafana.

Проект демонстрирует:

- gRPC-взаимодействие между сервисами
- Асинхронные события через Kafka (Transactional Outbox, idempotent consumers)
- PostgreSQL: транзакции, индексы, логическое шардирование / partitioning
- Observability: метрики, алерты и дашборды в Grafana

**Доменные события в Kafka**
- OrderCreated: “заказ создан” (order_id, user_id, сумма, items, timestamp)
- PaymentSucceeded: “оплата прошла” (order_id, payment_id, сумма, timestamp)
- PaymentFailed: “оплата не прошла” (order_id, причина)
- InventoryReserved: “товар зарезервирован” (order_id, sku->qty)
- InventoryReservationFailed: “не смогли зарезервировать”
- OrderCompleted / OrderCanceled: “заказ завершён/отменён”

[Protobuf contracts](https://github.com/ChernykhITMO/order-processing-proto)

## Архитектура

![architecture](docs/architecture.png)

## Конфигурация окружения

Локально используем один экземпляр Postgres и одного пользователя для всех сервисов, а базы разделяем по сервисам. Это самый простой и стабильный способ для демо-проекта.

1) Скопируйте шаблоны:
   ```bash
   cp .env.example .env
   cp orders/.env.example orders/.env
   cp payments/.env.example payments/.env
   ```

2) Первый запуск Postgres или смена учётных данных:
   ```bash
   docker compose -f docker-compose.yaml down -v
   docker compose -f docker-compose.yaml up -d
   ```

3) Создание баз во всех сервисах:
   ```bash
   make db-create
   ```

4) Запуск миграций во всех сервисах:
   ```bash
   make migrate-up
   ```
