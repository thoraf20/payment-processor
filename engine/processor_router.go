package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/thoraf20/payment-processor/model"
	"github.com/thoraf20/payment-processor/repository"
)

// ProcessorRouter directs payments to appropriate processors
type ProcessorRouter struct {
	processors       map[string]PaymentProcessor
	routingRules     []RoutingRule
	defaultProcessor string
	mu               sync.RWMutex
	Repo             repository.PaymentRepository
}

// RoutingRule defines criteria for processor selection
type RoutingRule struct {
	Name        string
	Condition   func(p *model.Payment) bool
	ProcessorID string
	Priority    int // Higher priority executes first
}

// NewProcessorRouter creates a configured router
func NewProcessorRouter() *ProcessorRouter {
	return &ProcessorRouter{
		processors: make(map[string]PaymentProcessor),
		routingRules: []RoutingRule{
			{
				Name:        "fallback",
				Condition:   func(_ *model.Payment) bool { return true },
				ProcessorID: "stripe",
				Priority:    0,
			},
		},
		defaultProcessor: "stripe",
	}
}

// RegisterProcessor adds a processor to the router
func (r *ProcessorRouter) RegisterProcessor(id string, processor PaymentProcessor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.processors[id]; exists {
		return fmt.Errorf("processor %q already registered", id)
	}

	r.processors[id] = processor
	return nil
}

// AddRoutingRule adds a new routing rule
func (r *ProcessorRouter) AddRoutingRule(rule RoutingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate processor exists
	if _, exists := r.processors[rule.ProcessorID]; !exists {
		return fmt.Errorf("processor %q not registered", rule.ProcessorID)
	}

	// Insert rule in priority order
	var inserted bool
	for i, existing := range r.routingRules {
		if rule.Priority > existing.Priority {
			r.routingRules = append(r.routingRules[:i], append([]RoutingRule{rule}, r.routingRules[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		r.routingRules = append(r.routingRules, rule)
	}

	return nil
}

// GetProcessor selects the appropriate processor
func (r *ProcessorRouter) GetProcessor(payment *model.Payment) (PaymentProcessor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check rules in priority order
	for _, rule := range r.routingRules {
		if rule.Condition(payment) {
			if processor, exists := r.processors[rule.ProcessorID]; exists {
				return processor, nil
			}
		}
	}

	// Fallback to default
	if defaultProc, exists := r.processors[r.defaultProcessor]; exists {
		return defaultProc, nil
	}

	return nil, errors.New("no suitable processor available")
}

// Implement PaymentProcessor interface by routing calls
func (r *ProcessorRouter) Authorize(ctx context.Context, payment *model.Payment) error {
	processor, err := r.GetProcessor(payment)
	if err != nil {
		return fmt.Errorf("processor selection failed: %w", err)
	}

	return processor.Authorize(ctx, payment)
}

func (r *ProcessorRouter) Capture(ctx context.Context, paymentID string, amount int64) error {
	// Need to get original payment to determine processor
	payment, err := r.Repo.Get(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	processor, err := r.GetProcessor(payment)
	if err != nil {
		return fmt.Errorf("processor selection failed: %w", err)
	}

	return processor.Capture(ctx, paymentID, amount)
}

func (r *ProcessorRouter) Refund(ctx context.Context, paymentID string, amount int64) error {
	payment, err := r.Repo.Get(ctx, paymentID)

	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	processor, err := r.GetProcessor(payment)
	if err != nil {
		return fmt.Errorf("processor selection failed: %w", err)
	}

	return processor.Refund(ctx, paymentID, amount)
}
