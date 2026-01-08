package payments_test

import (
	"bytes"
	"encoding/json"
	"errors"
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
		req, _ := http.NewRequest("GET", "/api/payments/test-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.NotNil(t, w.Body)
		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
	t.Run("PaymentNotFound", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/payments/NonExistingID", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

type ConfigurableBankGateway struct {
	ProcessPaymentFunc func(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error)
}

func (m *ConfigurableBankGateway) ProcessPayment(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error) {
	if m.ProcessPaymentFunc != nil {
		return m.ProcessPaymentFunc(req)
	}
	return &payments.BankAuthorization{}, nil
}

func TestPostPaymentHandler(t *testing.T) {
	validReq := payments.PostPaymentRequest{
		CardNumber:  "1234567890123456",
		ExpiryMonth: 12,
		ExpiryYear:  2030,
		Currency:    "USD",
		Amount:      1000,
		Cvv:         "123",
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		mockBankFunc   func(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Success: Payment Authorized",
			requestBody: validReq,
			mockBankFunc: func(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error) {
				return &payments.BankAuthorization{
					Authorized:        true,
					AuthorizationCode: "AUTH-123",
					ErrorMessage:      "",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"payment_status":        "Authorized",
				"card_number_last_four": "3456",
				"amount":                float64(1000),
				"currency":              "USD",
			},
		},
		{
			name:        "Success: Payment Declined",
			requestBody: validReq,
			mockBankFunc: func(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error) {
				return &payments.BankAuthorization{
					Authorized:        false,
					AuthorizationCode: "",
					ErrorMessage:      "Insufficient funds",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"payment_status":        "Declined",
				"card_number_last_four": "3456",
			},
		},
		{
			name: "Failure: Validation Error (Invalid Currency)",
			requestBody: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  2030,
				Currency:    "ZZZ",
				Amount:      1000,
				Cvv:         "123",
			},
			mockBankFunc:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"payment_status": "Rejected",
				"error_message":  "currency not supported",
			},
		},
		{
			name:           "Failure: Malformed JSON",
			requestBody:    "{{{{ not valid json",
			mockBankFunc:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"payment_status": "Rejected",
				"error_message":  "Invalid request body format",
			},
		},
		{
			name:        "Failure: Bank Service Unavailable",
			requestBody: validReq,
			mockBankFunc: func(req *payments.PostPaymentRequest) (*payments.BankAuthorization, error) {
				return nil, errors.New("bank timeout")
			},
			expectedStatus: http.StatusBadGateway,
			expectedBody: map[string]interface{}{
				"payment_status": "Failed",
				"error_message":  "Financial institution unavailable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := payments.NewPaymentsRepository()
			mockBank := &ConfigurableBankGateway{
				ProcessPaymentFunc: tt.mockBankFunc,
			}
			handler := payments.NewPaymentsHandler(storage, mockBank)

			var bodyBytes []byte
			if s, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}
			req, _ := http.NewRequest("POST", "/api/payments", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.PostHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var respBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &respBody)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, respBody[key], "Field %s mismatch", key)
			}

			if w.Code == http.StatusOK {
				id, ok := respBody["id"].(string)
				assert.True(t, ok, "Response should contain an ID")
				assert.NotEmpty(t, id)

				savedPayment := storage.GetPayment(id)
				assert.NotNil(t, savedPayment, "Payment should be saved in repository")
				assert.Equal(t, respBody["payment_status"], savedPayment.PaymentStatus)
			}
		})
	}
}
