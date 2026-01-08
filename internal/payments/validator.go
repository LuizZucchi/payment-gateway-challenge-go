package payments

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var allowedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"BRL": true,
}

var numericRegex = regexp.MustCompile(`^[0-9]+$`)

func (req *PostPaymentRequest) Validate() error {
	if err := req.validateCurrency(); err != nil {
		return err
	}
	if err := req.validateAmount(); err != nil {
		return err
	}
	if err := req.validateCardNumber(); err != nil {
		return err
	}
	if err := req.validateCVV(); err != nil {
		return err
	}
	return req.validateExpiry()
}

func (req *PostPaymentRequest) validateCurrency() error {
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	if !allowedCurrencies[req.Currency] {
		return errors.New("currency not supported")
	}
	return nil
}

func (req *PostPaymentRequest) validateAmount() error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

func (req *PostPaymentRequest) validateCardNumber() error {
	if req.CardNumber == "" {
		return errors.New("card_number is required")
	}

	cleanCard := strings.ReplaceAll(req.CardNumber, " ", "")

	if !numericRegex.MatchString(cleanCard) {
		return errors.New("card_number must contain only numeric characters")
	}

	if len(cleanCard) < 14 || len(cleanCard) > 19 {
		return errors.New("card_number must be between 14 and 19 characters")
	}
	return nil
}

func (req *PostPaymentRequest) validateCVV() error {
	if req.Cvv == "" {
		return errors.New("cvv is required")
	}
	if !numericRegex.MatchString(req.Cvv) {
		return errors.New("cvv must contain only numeric characters")
	}
	if len(req.Cvv) < 3 || len(req.Cvv) > 4 {
		return errors.New("cvv must be 3 or 4 characters")
	}
	return nil
}

func (req *PostPaymentRequest) validateExpiry() error {
	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return errors.New("expiry_month must be between 1 and 12")
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	if req.ExpiryYear < currentYear {
		return errors.New("expiry_year must be in the future")
	}

	if req.ExpiryYear == currentYear && req.ExpiryMonth < int(currentMonth) {
		return errors.New("expiry date must be in the future")
	}

	return nil
}
