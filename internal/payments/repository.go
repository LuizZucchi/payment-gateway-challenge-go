package payments

type getPaymentRequest struct {
	id       string
	respChan chan *PostPaymentResponse
}

type PaymentsRepository struct {
	addChan chan PostPaymentResponse
	getChan chan getPaymentRequest
}

func NewPaymentsRepository() *PaymentsRepository {
	repo := &PaymentsRepository{
		addChan: make(chan PostPaymentResponse),
		getChan: make(chan getPaymentRequest),
	}

	go repo.monitor()

	return repo
}

func (ps *PaymentsRepository) monitor() {
	var paymentsList []PostPaymentResponse

	for {
		select {
		case p := <-ps.addChan:
			paymentsList = append(paymentsList, p)

		case req := <-ps.getChan:
			var found *PostPaymentResponse
			for i := range paymentsList {
				if paymentsList[i].Id == req.id {
					clone := paymentsList[i]
					found = &clone
					break
				}
			}
			req.respChan <- found
		}
	}
}

func (ps *PaymentsRepository) GetPayment(id string) *PostPaymentResponse {
	respChan := make(chan *PostPaymentResponse)
	
	ps.getChan <- getPaymentRequest{
		id:       id,
		respChan: respChan,
	}

	return <-respChan
}

func (ps *PaymentsRepository) AddPayment(payment PostPaymentResponse) {
	ps.addChan <- payment
}