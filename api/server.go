// api/server.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thoraf20/payment-processor/engine"
	"github.com/thoraf20/payment-processor/model"
	"go.uber.org/zap"
)

type Server struct {
	router *mux.Router
	logger *zap.Logger
	paymentEngine *engine.PaymentEngine

}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}

func NewServer(logger *zap.Logger, paymentEngine *engine.PaymentEngine) *Server {
	r := mux.NewRouter()
	s := &Server{
		router:        r,
		logger:        logger,
		paymentEngine: paymentEngine,
	}
	
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/payments", s.handleCreatePayment()).Methods("POST")
	s.router.HandleFunc("/payments/{id}", s.handleGetPayment()).Methods("GET")
	s.router.HandleFunc("/payments/{id}/refund", s.handleRefund()).Methods("POST")
}

// Implement handlers using the paymentEngine
func (s *Server) handleCreatePayment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var payment model.Payment
		if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
			s.logger.Error("Failed to decode request", zap.Error(err))
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Process payment
		createdPayment, err := s.paymentEngine.CreatePayment(r.Context(), &payment)
		if err != nil {
			s.logger.Error("Failed to create payment", zap.Error(err))
			http.Error(w, "Payment processing failed", http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdPayment)
	}
}

func (s *Server) handleGetPayment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Payment logic
	}
}

func (s *Server) handleRefund() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Payment refund logic
	}
}
