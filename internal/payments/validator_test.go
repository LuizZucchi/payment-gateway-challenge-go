package payments_test

import (
	"testing"
	"time"

	"github.com/LuizZucchi/payment-gateway-challenge-go/internal/payments"
	"github.com/stretchr/testify/assert"
)

func TestPostPaymentRequest_Validate(t *testing.T) {
	futureYear := time.Now().Year() + 1

	tests := []struct {
		name    string
		req     payments.PostPaymentRequest
		wantErr string
	}{
		{
			name: "Valid Request (USD)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "USD",
				Amount:      1000,
				Cvv:         "123",
			},
			wantErr: "",
		},
		{
			name: "Valid Request (JPY)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "JPY",
				Amount:      1000,
				Cvv:         "123",
			},
			wantErr: "",
		},
		{
			name: "Invalid Currency (Not Supported)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "BRL",
				Amount:      100,
				Cvv:         "123",
			},
			wantErr: "currency not supported",
		},
		{
			name: "Invalid Currency (Empty)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "",
				Amount:      100,
				Cvv:         "123",
			},
			wantErr: "currency is required",
		},
		{
			name: "Invalid Amount (Zero)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "USD",
				Amount:      0,
				Cvv:         "123",
			},
			wantErr: "amount must be greater than 0",
		},
		{
			name: "Invalid Card Number (Empty)",
			req: payments.PostPaymentRequest{
				CardNumber:  "",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			wantErr: "card_number is required",
		},
		{
			name: "Invalid Card Number (Non-numeric)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234abcd",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			wantErr: "card_number must contain only numeric characters",
		},
		{
			name: "Invalid Card Number (Length)",
			req: payments.PostPaymentRequest{
				CardNumber:  "123",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "123",
			},
			wantErr: "card_number must be between 14 and 19 characters",
		},
		{
			name: "Invalid CVV (Empty)",
			req: payments.PostPaymentRequest{
				CardNumber:  "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  futureYear,
				Currency:    "USD",
				Amount:      100,
				Cvv:         "",
			},
			wantErr: "cvv is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}
