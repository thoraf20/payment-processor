package processors_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thoraf20/payment-processor/model"
	"github.com/thoraf20/payment-processor/processors"
	"github.com/thoraf20/payment-processor/repository/mock"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestFlutterwaveProcessor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPaymentRepository(ctrl)
	logger := zap.NewNop()

	processor := processors.NewFlutterwaveProcessor(
		"test_key",
		mockRepo,
		logger,
	)

	t.Run("Successful Authorization", func(t *testing.T) {
		payment := &model.Payment{
			ID:       "test_123",
			Amount:   1000, // NGN 10.00
			Currency: "NGN",
			PaymentMethod: model.PaymentMethod{
				Type: "card",
				Details: map[string]interface{}{
					"number":    "5531886652142950",
					"cvv":       "564",
					"exp_month": "09",
					"exp_year":  "32",
				},
			},
		}

		mockRepo.EXPECT().Save(gomock.Any(), payment).Return(nil).Times(2)
		
		err := processor.Authorize(context.Background(), payment)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusCompleted, payment.Status)
	})
}