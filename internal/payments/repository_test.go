package payments_test

import (
	"sync"
	"testing"
	"time"

	"github.com/LuizZucchi/payment-gateway-challenge-go/internal/payments"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddAndGetPayment(t *testing.T) {
	repo := payments.NewPaymentsRepository()
	id := uuid.New().String()
	
	inputPayment := payments.PostPaymentResponse{
		Id:                 id,
		PaymentStatus:      "Authorized",
		CardNumberLastFour: "1234",
		ExpiryMonth:        12,
		ExpiryYear:         2030,
		Currency:           "USD",
		Amount:             100,
	}

	repo.AddPayment(inputPayment)
	result := repo.GetPayment(id)

	assert.NotNil(t, result)
	if result != nil {
		assert.Equal(t, inputPayment.Id, result.Id)
		assert.Equal(t, inputPayment.PaymentStatus, result.PaymentStatus)
		assert.Equal(t, inputPayment.Amount, result.Amount)
	}
}

func TestGetPaymentNotFound(t *testing.T) {
	repo := payments.NewPaymentsRepository()
	nonExistentID := uuid.New().String()

	result := repo.GetPayment(nonExistentID)

	assert.Nil(t, result)
}

func TestConcurrencySafety(t *testing.T) {
	repo := payments.NewPaymentsRepository()
	concurrencyLevel := 100
	var wg sync.WaitGroup

	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			payment := payments.PostPaymentResponse{
				Id:            uuid.New().String(),
				PaymentStatus: "Authorized",
				Amount:        50,
			}
			repo.AddPayment(payment)
		}()
	}

	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomID := uuid.New().String()
			_ = repo.GetPayment(randomID)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		assert.FailNow(t, "Timeout")
	}
}