package payments_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LuizZucchi/payment-gateway-challenge-go/internal/payments"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type MockBankGateway struct{}

func (m *MockBankGateway) ProcessPayment(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error) {
	return &payments.BankAuthorization{}, nil
}

func TestGetPaymentHandler(t *testing.T) {
	payment := payments.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: "1234",
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}
	ps := payments.NewPaymentsRepository()
	ps.AddPayment(payment)

	handler := payments.NewPaymentsHandler(ps, &MockBankGateway{})

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", handler.GetHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	t.Run("PaymentFound", func(t *testing.T) {
		// Create a new HTTP request for testing
		req, _ := http.NewRequest("GET", "/api/payments/test-id", nil)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the body is not nil
		assert.NotNil(t, w.Body)

		// Check the HTTP status code in the response
		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
	t.Run("PaymentNotFound", func(t *testing.T) {
		// Create a new HTTP request for testing with a non-existing payment ID
		req, _ := http.NewRequest("GET", "/api/payments/NonExistingID", nil)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the HTTP status code in the response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}