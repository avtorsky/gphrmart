package accrual

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Accrualer interface {
	GetOrderInfo(orderID string) ([]byte, error)
}

type AccrualService struct {
	client *resty.Client
}

func NewAccrualService(accrualSystemAddr string) *AccrualService {
	client := resty.New()
	client.SetBaseURL(fmt.Sprintf("%v/api/orders/", accrualSystemAddr))
	return &AccrualService{client: client}
}

func (a *AccrualService) GetOrderInfo(orderID string) ([]byte, error) {
	res, err := a.client.R().Get(orderID)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() == http.StatusTooManyRequests {
		return nil, ErrTooManyRequests
	}
	return res.Body(), nil
}
