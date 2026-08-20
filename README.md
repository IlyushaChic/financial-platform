# 💰 Financial Platform

Pet-проект для демонстрации навыков разработки микросервисов на Go.

---

## 📖 О проекте

Финансовый сервис с микросервисной архитектурой, где пользователи могут регистрироваться, переводить деньги, получать уведомления и смотреть историю транзакций.

---

## 🧩 Технологии

- Go 1.26+, gRPC, GraphQL, WebSocket
- PostgreSQL, ClickHouse, Redis
- Kafka, RabbitMQ
- React, TypeScript, Tailwind, Apollo Client
- Docker, Kubernetes (minikube), Helm
- Prometheus, Grafana, Jaeger, Loki

---

## 🚀 Запуск локально

```bash
# Клонировать
git clone https://github.com/IlyushaChic/financial-platform.git
cd financial-platform

# Инфраструктура
cd backend
docker-compose up -d

# Миграции
make migrate-up

# Сервисы
make build
make run

# Фронтенд
cd ../frontend
npm install
npm run dev

# Проверка
Frontend: http://localhost:5173
GraphQL: http://localhost:8080/graphql
Grafana: http://localhost:3000 (admin/admin)
Jaeger: http://localhost:16686
RabbitMQ: http://localhost:15672 (admin/admin)

# Kubernetes (minikube)
# Запустить minikube
minikube start --driver=docker

# Установить Helm чарт
cd deployments/helm/financial-platform
helm install financial-platform . --namespace financial-platform --create-namespace

# Проверить
kubectl get pods -n financial-platform

# Добавить hosts
echo "$(minikube ip) financial-platform.local" | sudo tee -a /etc/hosts
Открыть: http://financial-platform.local



