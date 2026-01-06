package payments

type PostPaymentRequest struct {
	CardNumber         int    `json:"card_number"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
	Cvv                string `json:"cvv"`
}

type PostPaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"`
	CardNumberLastFour string `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}

type GetPaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"`
	CardNumberLastFour int    `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}
