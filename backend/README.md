# 💰 Financial Platform

> Пет-проект для демонстрации навыков разработки микросервисов на Go с использованием современных технологий.

---

## 🚀 Технологии

- **Go 1.22+**, gRPC, GraphQL, WebSocket  
- **PostgreSQL**, **Redis**, **ClickHouse**  
- **Kafka**, **RabbitMQ**  
- **Docker**, **Kubernetes**, **Helm**  
- **Prometheus**, **Grafana**, **Jaeger**, **Loki**  
- **React**, **TypeScript**, **Apollo Client**

---

## 🏗 Архитектура

```mermaid
flowchart LR
    Client[React SPA] --> Gateway[API Gateway]
    Gateway --> Auth[Auth Service]
    Gateway --> Tx[Transaction Service]
    Gateway --> Notif[Notification Service]

    Tx --> PG[(PostgreSQL)]
    Tx --> Redis[(Redis)]
    Tx --> Kafka[Kafka]
    Tx --> Rabbit[RabbitMQ]

    Kafka --> ClickHouse[(ClickHouse)]
    Kafka --> Notif

    Notif --> Rabbit
    Notif --> Providers[Email / SMS / Push]

    🐳 Запуск локально
1. Запусти инфраструктуру
bash
cd backend
docker-compose up -d
2. Примени миграции
bash
make migrate-up
3. Запусти сервисы
bash
make build
make run
4. Запусти фронтенд
bash
cd frontend
npm install
npm run dev
☸️ Развертывание в Kubernetes (minikube)
1. Запусти minikube
bash
minikube start --driver=docker
2. Установи Helm-чарт
bash
cd deployments/helm/financial-platform
helm install financial-platform . --namespace financial-platform --create-namespace
3. Проверь статус
bash
kubectl get pods -n financial-platform
📚 Документация
Подробное описание архитектуры и API доступно в docs/architecture.md.

📄 Лицензия
MIT © Ilya Chich

text

---

## ✅ Что получилось

- **Кратко** — всего несколько разделов.
- **Понятно** — любой разработчик поймёт, что это за проект и как его запустить.
- **Достаточно для портфолио** — есть схема, технологии, инструкции.

Ты можешь убрать или добавить разделы по своему усмотрению (например, убрать `mermaid`-схему, если она не отображается, или заменить текстовой схемой). Главное — не перегружать.

Удачи с проектом! Если хочешь, могу помочь с `docs/architecture.md` — тоже сделаем коротко и по делу. 😊