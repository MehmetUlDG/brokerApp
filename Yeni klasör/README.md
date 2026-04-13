# Broker Application Architecture

A high-performance, event-driven trading broker platform built with Go 1.22+ and microservices architecture. Designed using Clean Architecture principles, ensuring robust financial transactions and high availability.

![Go](https://img.shields.io/badge/go-1.22+-blue.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)
![Redis](https://img.shields.io/badge/Redis-7-red.svg)
![Kafka](https://img.shields.io/badge/Kafka-Event--Driven-black.svg)
![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)

## 🏗️ Architecture Overview

The platform uses a microservices architecture to ensure scalability, fault isolation, and independent deployability of distinct business domains:

*   **Broker Backend (`/broker-backend`)**: The core gateway, handling user authentication, order routing, wallet management, and API endpoints. 
    *   **Wallet Service**: Handles financial data with strict precision (`DECIMAL(18,8)`) and guarantees transaction safety using PostgreSQL pessimistic locking (`SELECT ... FOR UPDATE`).
    *   **Order Service**: Manages order lifecycles and utilizes the **Outbox Pattern** to ensure eventual consistency between the PostgreSQL database and Kafka message queue without distributed transactions.
*   **Matching Engine (`/matching-engine`)**: A high-speed, low-latency engine built to match buy and sell orders accurately and emit trade execution events.
*   **Ingestion Service (`/ingestion-service`)**: Fetches and processes live market data, feeding the platform with real-time price tick information via Kafka topics.

## 🚀 Key Technologies & Patterns

*   **Go (Golang)**: Chosen for its concurrency model and high performance.
*   **Clean Architecture**: Separation of concerns (Domain, Use Cases, Repository, Delivery) for maintainability.
*   **Event-Driven Architecture**: Apache Kafka serves as the central nervous system, decoupling services and managing high-throughput event streaming.
*   **Outbox Pattern**: Ensuring atomicity between database writes and event publishing.
*   **Data Integrity (`sqlx`)**: Direct, optimized SQL queries using `sqlx` instead of ORMs like GORM for maximum performance and precise transaction control.
*   **Caching**: Redis for session management and low-latency data access.

## 📁 Project Structure

```text
.
├── broker-backend/       # Core API, Wallet & Order domains, Docker Compose infra
├── ingestion-service/    # Market data fetching and publishing
└── matching-engine/      # Order book management and trade execution
```

## 🛠️ Getting Started

### Prerequisites

*   [Docker](https://www.docker.com/) and Docker Compose
*   [Go 1.22+](https://go.dev/)
*   `make` (optional, for running scripts easily)

### Running the Infrastructure locally

The foundation of the platform runs via Docker Compose in the `broker-backend` directory. This includes PostgreSQL, Redis, Kafka, and Zookeeper with properly configured health checks.

```bash
cd broker-backend

# Copy environment template
cp .env.example .env

# Start all infrastructure dependencies
docker-compose up -d

# Check the status of the containers
docker-compose ps
```

## 🤝 Contributing

This is a professional-grade template. All PRs should ensure test coverage, clean architecture boundaries, and proper documentation of domain changes.

---
*Built with focus on performance, ACID compliance, and system reliability.*
