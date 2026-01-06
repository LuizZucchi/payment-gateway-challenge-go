package payments

type PaymentsRepository struct {
	payments []PostPaymentResponse
}

func NewPaymentsRepository() *PaymentsRepository {
	return &PaymentsRepository{
		payments: []PostPaymentResponse{},
	}
}

func (ps *PaymentsRepository) GetPayment(id string) *PostPaymentResponse {
	for _, element := range ps.payments {
		if element.Id == id {
			return &element
		}
	}
	return nil
}

func (ps *PaymentsRepository) AddPayment(payment PostPaymentResponse) {
	ps.payments = append(ps.payments, payment)
}
