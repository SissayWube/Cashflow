# Cashflow Payment Gateway

A production-ready payment gateway module built in Go, demonstrating reliable asynchronous processing, idempotency, and concurrency safety.

## Overview

This project implements a minimal but realistic payment gateway with the following key properties:

* Asynchronous processing via RabbitMQ
* Idempotent payment processing – a payment is processed exactly once, even under message redelivery, retries, or concurrent workers
* PostgreSQL as the source of truth with row-level locking and transactions
* Clean separation of API and worker services
* Containerized with Docker Compose

## Core Features

### 1. Payment Creation API

**Endpoint:** `POST /payments`

```json
{
  "amount": 150.00,
  "currency": "USD",
  "reference": "order-12345"
}
```

* Validates input (amount > 0, currency ETB or USD, reference required and unique)
* Inserts a record with status `PENDING`
* Publishes the payment ID to the RabbitMQ queue `payments`
* Returns the created payment with ID and status

### 2. Asynchronous Processing (Worker)

* Consumes messages from RabbitMQ
* Uses `SELECT ... FOR UPDATE` to lock the row
* Processes only if status is `PENDING`
* Simulates processing (2s delay) and randomly sets status to `SUCCESS` (70%) or `FAILED` (30%)
* Skips processing for terminal states → guarantees idempotency

### 3. Payment Status Retrieval

**Endpoint:** `GET /payments/:id`

Returns full payment details including current status and creation timestamp.

## Project Structure

```text
Cashflow/
├── api/                  # API service
│   ├── main.go
│   ├── api.go
│   ├── db.go
│   ├── mq.go
│   ├── model.go
│   └── Dockerfile
├── worker/               # Background worker
│   ├── main.go
│   ├── db.go
│   ├── mq.go
│   └── Dockerfile
├── docker-compose.yml    # Orchestrates db, rabbitmq, api, worker
├── .env                  # Local development environment variables (optional in Docker)
└── README.md             # This file
```

## Tech Stack

* **API:** Go + Echo framework
* **Worker:** Go + RabbitMQ consumer (`streadway/amqp`)
* **Database:** PostgreSQL 15
* **Messaging:** RabbitMQ 3-management
* **Migrations:** Embedded via golang-migrate with `iofs`
* **Containerization:** Docker + Docker Compose

## Getting Started

### Prerequisites

* Docker Desktop (Compose v2)
* Git

### Run the System

```bash
git clone <your-repo>
cd Cashflow
docker compose up --build
```

**Services:**

* API → [http://localhost:8080](http://localhost:8080)
* RabbitMQ Management → [http://localhost:15672](http://localhost:15672) (guest/guest)
* PostgreSQL → localhost:5432

## Test the Flow

### Create a Payment

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{"amount":100,"currency":"USD","reference":"test-001"}'
```

### Check Status

```bash
curl http://localhost:8080/payments/<id>
```

The status will eventually become `SUCCESS` or `FAILED`.

## Verify Idempotency & Redelivery

1. Let a payment be processed normally.
2. In RabbitMQ UI → **Queues** → `payments`, publish a message with the same payment ID.
3. Worker logs will show `Skipping payment X (already ...)` and the status remains unchanged.

## Environment Variables

For local development without Docker, create a `.env` file:

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=dbpass
DB_NAME=cashflow

MQ_HOST=localhost
MQ_USER=guest
MQ_PASSWORD=guest
MQ_QUEUE=payments

APP_ENV=development
```

In containers, `APP_ENV=production` is set to skip `.env` loading.

## Scaling Workers

Edit `docker-compose.yml`:

```yaml
worker:
  scale: 4
```

Multiple workers compete safely thanks to row-level locking.

## Stopping

```bash
docker compose down
```

```bash
docker compose down -v
```

## Conclusion

This project demonstrates a real-world pattern for building reliable payment (or any stateful) workflows:

* API accepts requests and persists state
* Message queue drives asynchronous work
* Worker ensures exactly-once semantics using database transactions and locking
* All services are containerized and orchestrated

Feel free to extend this project with webhooks, retries, monitoring, or real payment provider integration.
