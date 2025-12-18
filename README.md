# order-processing-platform
Система обработки заказов с gRPC-взаимодействием, событийной коммуникацией через Kafka, PostgreSQL и полноценным мониторингом в Prometheus + Grafana.

Микросервисная платформа обработки заказов.
Проект демонстрирует:

- gRPC-взаимодействие между сервисами
- Асинхронные события через Kafka (Transactional Outbox, idempotent consumers)
- PostgreSQL: транзакции, индексы, логическое шардирование / partitioning
- Observability: метрики, алерты и дашборды в Grafana
