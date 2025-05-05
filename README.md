📦# Payment Processing Engine

![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-336791?logo=postgresql)
![Stripe](https://img.shields.io/badge/Stripe-API-008CDD?logo=stripe)

A production-ready payment processing system built with Go that supports multiple payment providers (Stripe, Flutterwave, Paystack) with clean architecture and robust error handling.

## Architecture Overview

```mermaid
graph TD
    A[API Layer] --> B[Payment Engine]
    B --> C[Processors]
    B --> D[Repository]
    D --> E[(PostgreSQL)]
    C --> F[Stripe]
    C --> G[FlutterWave]
    C --> H[PayStack]

## Architecture
┌─────────────────────────────────────────────────┐
│                 Payment Engine                  │
├─────────────────┬───────────────┬───────────────┤
│  API Layer      │  Core Engine  │  Data Layer   │
│  (HTTP/gRPC)    │               │  (PostgreSQL) │
└─────────────────┴───────────────┴───────────────┘

Key Features
💳 Multi-provider support (Stripe, FlutterWave, PayStack)

🔐 PCI-compliant payment handling

📊 Built-in metrics and observability

🔄 Idempotent operation support

🛡️ Graceful shutdown and recovery

📝 Structured logging with Zap

⚙️ Environment-based configuration

# Getting Started
## Prerequisites

Go 1.20+

PostgreSQL 13+

Payment provider accounts (Stripe/FlutterWave/PayStack)

Installation

```bash
git clone https://github.com/thoraf20/payment-processor.git
cd payment-processor
cp .env.example .env
# Edit .env with your credentials
```

Configuration

# Required
DATABASE_URL=postgres://user:password@localhost:5432/payments
STRIPE_API_KEY=sk_test_your_key
FLUTTERWAVE_API_KEY=FLWSECK_TEST_your_key
PAYSTACK_API_KEY=sk_test_your_key

# Optional (defaults shown)
HTTP_PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info


## Project Structure

Directory                               Purpose
--------------------------------------------------------------------------
/api	          HTTP handlers and middleware

/cmd	          Main application entry points

/config	        Environment configuration loading

/engine	        Core payment processing business logic

/model	        Domain models and data structures

/processors	    Payment provider integrations (Stripe, Flutterwave, etc.)

/repository	    Database access layer

/logger	        Logging configuration and utilities

## API Endpoints

POST   /payments          - Create new payment
GET    /payments/{id}     - Retrieve payment
POST   /payments/{id}/capture - Capture authorized payment
POST   /payments/{id}/refund  - Process refund

# System

GET    /health       - Service health check
GET    /metrics      - Prometheus metrics

## Running the Service

go run cmd/main.go

## Production (Docker)

docker-compose up --build

## Operational Features

# Monitoring

Built-in Prometheus metrics available at /metrics:

  payment_requests_total

  payment_processing_time_seconds

  database_operations_total

# Logging

{
  "level": "info",
  "ts": "2023-07-15T10:00:00Z",
  "caller": "api/server.go:45",
  "msg": "Payment processed",
  "payment_id": "pay_123",
  "amount": 1000,
  "currency": "USD",
  "duration_ms": 125
}

# Error Handling

Three-tier error handling strategy:

  Domain errors - Business logic failures

  Infrastructure errors - Database/network issues

  Provider errors - Payment processor failures

API Documentation

Create a Payment

POST /payments
Content-Type: application/json

{
  "amount": 1000,
  "currency": "usd",
  "payment_method": {
    "type": "card",
    "card": {
      "number": "4242424242424242",
      "exp_month": 12,
      "exp_year": 2025,
      "cvc": "123"
    }
  }
}

Response:

HTTP/1.1 201 Created
Content-Type: application/json

{
  "id": "pay_123456789",
  "amount": 1000,
  "currency": "usd",
  "status": "completed",
  "created_at": "2023-01-01T00:00:00Z"
}

Get Payment Details

GET /payments/{id}

Process Refund

POST /payments/{id}/refund
Content-Type: application/json

{
  "amount": 1000
}


Additional Production Considerations
Idempotency: Implement idempotency keys for payment requests

Retry Logic: For transient failures with payment processors

Webhooks: For async payment status updates

Circuit Breakers: For external service calls

Rate Limiting: To protect against abuse

Data Encryption: For sensitive payment data

Compliance: PCI DSS compliance considerations

Multi-Processor Support: Fallback processors

Batch Processing: For settlements and reconciliations

Audit Logging: For all payment operations
