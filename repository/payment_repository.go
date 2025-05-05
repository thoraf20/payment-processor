package repository

import (
	"context"
	"database/sql"

	"github.com/thoraf20/payment-processor/model"

	"go.uber.org/zap"
)

type PaymentFilter struct {
	Status    string
	Currency  string
	StartDate string
	EndDate   string
}

type PaymentRepository interface {
	Save(ctx context.Context, payment *model.Payment) error
	Get(ctx context.Context, id string) (*model.Payment, error)
	List(ctx context.Context, filter PaymentFilter) ([]*model.Payment, error)
}

type DbPaymentRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPaymentRepository(db *sql.DB, logger *zap.Logger) *DbPaymentRepository {
	return &DbPaymentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *DbPaymentRepository) Save(ctx context.Context, payment *model.Payment) error {
	// Your implementation here
	query := `INSERT INTO payments (id, external_id, amount, currency, status, payment_method_type, 
	          payment_method_details, created_at, updated_at, metadata)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	          ON CONFLICT (id) DO UPDATE SET
	          external_id = $2, amount = $3, currency = $4, status = $5,
	          payment_method_type = $6, payment_method_details = $7,
	          updated_at = $9, metadata = $10`
	
	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.ExternalID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod.Type,
		payment.PaymentMethod.Details,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.Metadata,
	)
	return err
}

func (r *DbPaymentRepository) Get(ctx context.Context, id string) (*model.Payment, error) {
	query := `SELECT id, external_id, amount, currency, status, payment_method_type, 
	         payment_method_details, created_at, updated_at, metadata
	         FROM payments WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	
	var payment model.Payment
	var details []byte // Assuming JSON storage for payment method details
	
	err := row.Scan(
		&payment.ID,
		&payment.ExternalID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod.Type,
		&details,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.Metadata,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or return a custom "not found" error
		}
		return nil, err
	}
	
	// Convert details bytes to map[string]interface{}
	// You'll need to implement this based on how you store the details
	// payment.PaymentMethod.Details = parseDetails(details)
	
	return &payment, nil
}

func (r *DbPaymentRepository) List(ctx context.Context, filter PaymentFilter) ([]*model.Payment, error) {
	// Implement your listing logic here
	return nil, nil
}