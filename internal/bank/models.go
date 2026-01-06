package bank

type BankPaymentRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"` // Format: "MM/YYYY"
	Currency   string `json:"currency"`
	Amount     int    `json:"amount"`
	Cvv        string `json:"cvv"`
}

type BankPaymentResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
	ErrorMessage      string `json:"error_message,omitempty"`
}
