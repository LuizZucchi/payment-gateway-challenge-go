package payments

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// BankAuthorization define o que o Gateway espera receber do Banco.
// Definimos aqui para não depender do pacote 'bank'.
type BankAuthorization struct {
	Authorized        bool
	AuthorizationCode string
	ErrorMessage      string
}

// BankGateway define o contrato que qualquer cliente bancário deve seguir.
type BankGateway interface {
	ProcessPayment(req *PostPaymentRequest) (*BankAuthorization, error)
}

type PaymentsHandler struct {
	storage    *PaymentsRepository
	bankClient BankGateway
}

func NewPaymentsHandler(storage *PaymentsRepository, bankClient BankGateway) *PaymentsHandler {
	return &PaymentsHandler{
		storage:    storage,
		bankClient: bankClient,
	}
}

// GetHandler returns an http.HandlerFunc that handles HTTP GET requests.
// It retrieves a payment record by its ID from the storage.
// The ID is expected to be part of the URL.
func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		payment := h.storage.GetPayment(id)

		if payment != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(payment); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (h *PaymentsHandler) PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PostPaymentRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error_message":  "Invalid request body format",
				"payment_status": "Rejected",
			})
			return
		}

		if err := req.Validate(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error_message":  err.Error(),
				"payment_status": "Rejected",
			})
			return
		}
		bankResponse, err := h.bankClient.ProcessPayment(&req)
		if err != nil {
			h.respondWithError(w, http.StatusBadGateway, "Financial institution unavailable", "Failed")
			return
		}

		status := "Declined"
		if bankResponse.Authorized {
			status = "Authorized"
		}

		paymentID := uuid.New().String()
		lastFour := req.CardNumber[len(req.CardNumber)-4:]

		response := PostPaymentResponse{
			Id:                 paymentID,
			PaymentStatus:      status,
			CardNumberLastFour: lastFour,
			ExpiryMonth:        req.ExpiryMonth,
			ExpiryYear:         req.ExpiryYear,
			Currency:           req.Currency,
			Amount:             req.Amount,
		}

		h.storage.AddPayment(response)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func (h *PaymentsHandler) respondWithError(w http.ResponseWriter, code int, msg string, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error_message":  msg,
		"payment_status": status,
	})
}