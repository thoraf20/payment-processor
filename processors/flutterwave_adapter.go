package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/thoraf20/payment-processor/model"
	"github.com/thoraf20/payment-processor/repository"
	"go.uber.org/zap"
)

const (
	flutterwaveBaseURL = "https://api.flutterwave.com/v3"
	timeout            = 15 * time.Second
)

type FlutterwaveProcessor struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	repo       repository.PaymentRepository
	logger     *zap.Logger
}

// Helper function to convert map[string]string to map[string]interface{}
func convertToMapInterface(input map[string]string) map[string]interface{} {
	converted := make(map[string]interface{})
	for key, value := range input {
		converted[key] = value
	}
	return converted
}

func NewFlutterwaveProcessor(apiKey string, repo repository.PaymentRepository, logger *zap.Logger) *FlutterwaveProcessor {
	return &FlutterwaveProcessor{
		apiKey:  apiKey,
		baseURL: flutterwaveBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		repo:   repo,
		logger: logger.With(zap.String("processor", "flutterwave")),
	}
}

// Flutterwave API Request/Response Types
type flutterwaveChargeRequest struct {
	Amount     float64                `json:"amount"`
	Currency   string                 `json:"currency"`
	Email      string                 `json:"email"`
	TxRef      string                 `json:"tx_ref"`
	PaymentType string                `json:"payment_type"`
	Meta       map[string]interface{} `json:"meta"`
	Card       flutterwaveCardDetails `json:"card,omitempty"`
}

type flutterwaveCardDetails struct {
	Number      string `json:"card_number"`
	Cvv         string `json:"cvv"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
}

type flutterwaveResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID           int    `json:"id"`
		TxRef        string `json:"tx_ref"`
		Status       string `json:"status"`
		Processor    string `json:"processor_response"`
		AuthModel    string `json:"auth_model"`
		Currency     string `json:"currency"`
		Amount       float64 `json:"amount"`
		RedirectURL  string `json:"redirect_url"`
	} `json:"data"`
}

func (f *FlutterwaveProcessor) Authorize(ctx context.Context, payment *model.Payment) error {

	// Generate unique transaction reference
	txRef := fmt.Sprintf("flw-%s-%d", payment.ID, time.Now().Unix())

	reqBody := flutterwaveChargeRequest{
		Amount:     float64(payment.Amount) / 100, // Convert to currency unit
		Currency:   payment.Currency,
		TxRef:      txRef,
		PaymentType: "card",
		Meta: convertToMapInterface(payment.Metadata),
	}

	// Add card details if present
	if payment.PaymentMethod.Type == "card" {
		details := payment.PaymentMethod.Details
			reqBody.Card = flutterwaveCardDetails{
				Number:      details["number"].(string),
				Cvv:         details["cvv"].(string),
				ExpiryMonth: details["exp_month"].(string),
				ExpiryYear:  details["exp_year"].(string),
			}
	}

	// Save initial payment state with Flutterwave reference
	payment.ExternalID = txRef
	if err := f.repo.Save(ctx, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	resp, err := f.makeRequest(ctx, "/charges?type=card", reqBody)
	if err != nil {
		return fmt.Errorf("flutterwave API error: %w", err)
	}

	switch resp.Data.Status {
	case "successful":
		payment.Status = model.StatusCompleted
	case "pending":
		payment.Status = model.StatusPending
	default:
		payment.Status = model.StatusFailed
		return fmt.Errorf("payment failed: %s", resp.Data.Processor)
	}

	// Update payment with processor response
	return f.repo.Save(ctx, payment)
}

func (f *FlutterwaveProcessor) Capture(ctx context.Context, paymentID string, amount int64) error {
	// Flutterwave typically captures automatically, this validates the capture
	payment, err := f.repo.Get(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// external call to flutterwave to verify transaction
	resp, err := f.makeRequest(ctx, fmt.Sprintf("/transactions/%s/verify", payment.ExternalID), nil)
	if err != nil {
		return err
	}

	if resp.Data.Status != "successful" {
		return fmt.Errorf("cannot capture - transaction status: %s", resp.Data.Status)
	}

	payment.Status = model.StatusCompleted
	return f.repo.Save(ctx, payment)
}

func (f *FlutterwaveProcessor) Refund(ctx context.Context, paymentID string, amount int64) error {
	payment, err := f.repo.Get(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	req := struct {
		Amount int64 `json:"amount"`
	}{
		Amount: amount,
	}

	resp, err := f.makeRequest(ctx, fmt.Sprintf("/transactions/%s/refund", payment.ExternalID), req)
	if err != nil {
		return err
	}

	if resp.Status != "success" {
		return fmt.Errorf("refund failed: %s", resp.Message)
	}

	payment.Status = model.StatusRefunded
	return f.repo.Save(ctx, payment)
}

func (f *FlutterwaveProcessor) makeRequest(ctx context.Context, path string, body interface{}) (*flutterwaveResponse, error) {
	url := f.baseURL + path

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.apiKey)

	f.logger.Debug("Making request to Flutterwave",
		zap.String("url", url),
		zap.Any("request", body),
	)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("flutterwave error: %s", errorResp.Message)
	}

	var result flutterwaveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}