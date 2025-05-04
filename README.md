📦Payment Engine

A production-grade payment processing engine built with Go, designed for reliability, scalability, and security.

## Features
Multi-Processor Support: Integrates with Stripe, PayStack and FlutterWave (extensible to other providers).

REST API: Clean interface for payment operations

Idempotent Operations: Safe retries for failed requests

Comprehensive Metrics: Prometheus instrumentation

Structured Logging: Zap logger integration

Configurable: Environment variable based configuration

Database Backed: PostgreSQL storage for payment records

## Architecture
┌─────────────────────────────────────────────────┐
│                 Payment Engine                  │
├─────────────────┬───────────────┬───────────────┤
│  API Layer      │  Core Engine  │  Data Layer   │
│  (HTTP/gRPC)    │               │  (PostgreSQL) │
└─────────────────┴───────────────┴───────────────┘

# Getting Started
## Prerequisites
Go 1.20+

PostgreSQL 12+

Stripe account, PayStack and FlutterWave (for payment processing)

Installation

1. git clone https://github.com/thoraf20/payment-processor.git
cd payment-engine

2. Set up environment variables:
cp .env.example .env
# Edit .env with your configuration

3. Install dependencies:
go mod download

4. Run database migrations:

Running the Service

go run cmd/main.go

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


// Initialize repositories
	// paymentRepo := repository.NewPostgresPaymentRepository(db, log)
	
	// // Initialize payment processor
	// stripeProcessor := processors.NewStripeProcessor(cfg.StripeAPIKey)
	
	// // Initialize payment engine
	// paymentEngine := engine.NewPaymentEngine(stripeProcessor, paymentRepo)
	
	// // Initialize HTTP server
	// server := api.NewServer(log, paymentEngine)

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



processor := processors.NewStripeProcessor("your_stripe_secret_key")

payment := &engine.Payment{
    Amount:   1000, // $10.00
    Currency: "usd",
    PaymentMethod: engine.PaymentMethod{
        Type: "card",
        Details: map[string]interface{}{
            "number":    "4242424242424242",
            "exp_month": 12,
            "exp_year":  2025,
            "cvc":       "123",
        },
    },
}

err := processor.Authorize(context.Background(), payment)
if err != nil {
    // Handle error
    fmt.Printf("Payment failed: %v\n", err)
}