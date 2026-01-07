package bank_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LuizZucchi/payment-gateway-challenge-go/internal/payments"
	// Ajuste o import abaixo para o caminho correto do seu pacote bank
	"github.com/LuizZucchi/payment-gateway-challenge-go/internal/bank"
)

func TestBankClient_ProcessPayment(t *testing.T) {
	tests := []struct {
		name          string
		inputRequest  *payments.PostPaymentRequest
		mockHandler   func(w http.ResponseWriter, r *http.Request)
		expectedResp  *bank.BankPaymentResponse
		expectedError error
		errorContains string
	}{
		{
			name: "Success: Payment Authorized (200 OK)",
			inputRequest: &payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 10,
				ExpiryYear:  2028,
				Currency:    "BRL",
				Amount:      10050,
				Cvv:         "123",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/payments" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var receivedReq bank.BankPaymentRequest
				if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "invalid json sent by client"}`))
					return
				}

				if receivedReq.ExpiryDate != "10/2028" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf(`{"error": "wrong date format", "got": "%s"}`, receivedReq.ExpiryDate)))
					return
				}
				if receivedReq.Amount != 10050 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "wrong amount"}`))
					return
				}

				resp := bank.BankPaymentResponse{
					Authorized:        true,
					AuthorizationCode: "AUTH-12345",
					ErrorMessage:      "",
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &bank.BankPaymentResponse{
				Authorized:        true,
				AuthorizationCode: "AUTH-12345",
			},
			expectedError: nil,
		},
		{
			name: "Success: Payment Declined by Bank Logic (200 OK)",
			inputRequest: &payments.PostPaymentRequest{
				CardNumber: "0000000000000000",
				Amount:     100,
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := bank.BankPaymentResponse{
					Authorized:   false,
					ErrorMessage: "insufficient funds",
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &bank.BankPaymentResponse{
				Authorized:   false,
				ErrorMessage: "insufficient funds",
			},
			expectedError: nil,
		},
		{
			name:         "Failure: 400 Bad Request (Invalid Data)",
			inputRequest: &payments.PostPaymentRequest{Amount: 0},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"code": "invalid_parameter", "message": "amount must be positive"}`))
			},
			expectedResp:  nil,
			errorContains: "bank rejected request (400)",
		},
		{
			name:         "Failure: 503 Service Unavailable",
			inputRequest: &payments.PostPaymentRequest{Amount: 500},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedResp:  nil,
			expectedError: bank.ErrBankUnavailable,
		},
		{
			name:         "Failure: Invalid Response JSON from Bank",
			inputRequest: &payments.PostPaymentRequest{Amount: 500},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{not-a-valid-json`))
			},
			expectedResp:  nil,
			errorContains: "failed to decode bank response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			client := bank.NewBankClient(server.URL)

			resp, err := client.ProcessPayment(tt.inputRequest)

			if tt.expectedError != nil {
				if err != tt.expectedError {
					t.Errorf("Expected error target '%v', got '%v'", tt.expectedError, err)
				}
			} else if tt.errorContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%v'", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.expectedResp != nil {
				if resp == nil {
					t.Fatal("Expected response object, got nil")
				}
				if resp.Authorized != tt.expectedResp.Authorized {
					t.Errorf("Expected Authorized=%v, got %v", tt.expectedResp.Authorized, resp.Authorized)
				}
				if resp.AuthorizationCode != tt.expectedResp.AuthorizationCode {
					t.Errorf("Expected AuthCode=%s, got %s", tt.expectedResp.AuthorizationCode, resp.AuthorizationCode)
				}
			} else if resp != nil {
				t.Error("Expected nil response, got object")
			}
		})
	}
}
