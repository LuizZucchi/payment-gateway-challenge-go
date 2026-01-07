package bank

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/LuizZucchi/payment-gateway-challenge-go/internal/payments"
)

var ErrBankUnavailable = errors.New("bank service is unavailable")

type Client interface {
	ProcessPayment(req *payments.PostPaymentRequest) (*BankPaymentResponse, error)
}

type BankClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewBankClient(baseURL string) *BankClient {
	return &BankClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (bc *BankClient) ProcessPayment(req *payments.PostPaymentRequest) (*BankPaymentResponse, error) {
	bankReq := BankPaymentRequest{
		Amount:     req.Amount,
		Currency:   req.Currency,
		CardNumber: req.CardNumber,
		ExpiryDate: bc.formatExpiryDate(req.ExpiryMonth, req.ExpiryYear),
		Cvv:        req.Cvv,
	}

	requestBody, err := json.Marshal(bankReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bank request: %w", err)
	}

	url := fmt.Sprintf("%s/payments", bc.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create bank request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := bc.httpClient.Do(httpReq)
	if err != nil {
		return nil, ErrBankUnavailable
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var bankResp BankPaymentResponse
		if err := json.NewDecoder(resp.Body).Decode(&bankResp); err != nil {
			return nil, fmt.Errorf("failed to decode bank response: %w", err)
		}
		return &bankResp, nil
	case http.StatusBadRequest:
		var errorResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("bank rejected request (400): %v", errorResp)

	case http.StatusServiceUnavailable:
		return nil, ErrBankUnavailable

	default:
		return nil, fmt.Errorf("unexpected status code from bank: %d", resp.StatusCode)
	}
}

func (c *BankClient) formatExpiryDate(month, year int) string {
	return fmt.Sprintf("%02d/%d", month, year) // MM/YYYY format
}
