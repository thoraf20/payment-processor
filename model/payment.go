// model/payment.go
package model

import "time"

type PaymentStatus string

const (
	StatusPending    PaymentStatus = "pending"
	StatusAuthorized PaymentStatus = "authorized"
	StatusCompleted  PaymentStatus = "completed"
	StatusFailed     PaymentStatus = "failed"
	StatusRefunded   PaymentStatus = "refunded"
)

type Payment struct {
	ID            string
	ExternalID    string
	Amount        int64
	Currency      string
	Status        PaymentStatus
	PaymentMethod PaymentMethod
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Metadata      map[string]string
}

type PaymentMethod struct {
	Type    string
	Details map[string]interface{}
}