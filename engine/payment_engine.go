// engine/payment_engine.go
package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thoraf20/payment-processor/model"
	"github.com/thoraf20/payment-processor/repository"
)

type PaymentProcessor interface {
	Authorize(ctx context.Context, payment *model.Payment) error
	Capture(ctx context.Context, paymentID string, amount int64) error
	Refund(ctx context.Context, paymentID string, amount int64) error
}

type PaymentEngine struct {
	processor PaymentProcessor
	repo      repository.PaymentRepository
}

func NewPaymentEngine(processor PaymentProcessor, repo repository.PaymentRepository) *PaymentEngine {
	return &PaymentEngine{
		processor: processor,
		repo:      repo,
	}
}

func (e *PaymentEngine) CreatePayment(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	payment.ID = uuid.New().String()
	payment.Status = model.StatusPending
	payment.CreatedAt = time.Now().UTC()
	payment.UpdatedAt = payment.CreatedAt
	
	if err := e.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save initial payment: %w", err)
	}
	
	if err := e.processor.Authorize(ctx, payment); err != nil {
		payment.Status = model.StatusFailed
		_ = e.repo.Save(ctx, payment)
		return nil, fmt.Errorf("authorization failed: %w", err)
	}
	
	if payment.Status == model.StatusAuthorized {
		return payment, nil
	}
	
	payment.Status = model.StatusCompleted
	if err := e.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save completed payment: %w", err)
	}
	
	return payment, nil
}

// Implement other methods...