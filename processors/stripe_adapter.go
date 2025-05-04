// processors/stripe_adapter.go
package processors

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/thoraf20/payment-processor/model"
)

type StripeProcessor struct {
	apiKey string
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	stripe.Key = apiKey
	return &StripeProcessor{apiKey: apiKey}
}

func (s *StripeProcessor) Authorize(ctx context.Context, payment *model.Payment) error {
	// First create a PaymentMethod
	pmParams := &stripe.PaymentMethodParams{
		Type: stripe.String("card"),
		Card: &stripe.PaymentMethodCardParams{
			Number:   stripe.String(payment.PaymentMethod.Details["number"].(string)),
			ExpMonth: stripe.String(fmt.Sprintf("%v", payment.PaymentMethod.Details["exp_month"])),
			ExpYear:  stripe.String(fmt.Sprintf("%v", payment.PaymentMethod.Details["exp_year"])),
			CVC:      stripe.String(payment.PaymentMethod.Details["cvc"].(string)),
		},
	}

	pm, err := paymentmethod.New(pmParams)
	if err != nil {
		return fmt.Errorf("failed to create payment method: %w", err)
	}

	// Then create and confirm the PaymentIntent
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(payment.Amount),
		Currency:      stripe.String(payment.Currency),
		PaymentMethod: stripe.String(pm.ID),
		Confirm:       stripe.Bool(true),
	}

	_, err = paymentintent.New(params)
	if err != nil {
		return fmt.Errorf("failed to create payment intent: %w", err)
	}

	return nil
}

func (s *StripeProcessor) Capture(ctx context.Context, paymentID string, amount int64) error {
	params := &stripe.PaymentIntentCaptureParams{
		AmountToCapture: stripe.Int64(amount),
	}
	_, err := paymentintent.Capture(paymentID, params)
	return err
}

func (s *StripeProcessor) Refund(ctx context.Context, paymentID string, amount int64) error {
	// First get the PaymentIntent to check its status
	pi, err := paymentintent.Get(paymentID, nil)
	if err != nil {
		return fmt.Errorf("failed to get payment intent: %w", err)
	}

	// Only allow refunds on succeeded payments
	if pi.Status != stripe.PaymentIntentStatusSucceeded {
		return errors.New("can only refund succeeded payments")
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentID),
		Amount:        stripe.Int64(amount),
	}
	_, err = refund.New(params)
	return err
}