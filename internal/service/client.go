package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"

	getAccrualPath = "/api/orders/"
)

type AccrualClient interface {
	GetAccrual(orderNumber string) (AccrualOrder, error)
}

type AccrualOrder struct {
	Number  string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

type AccrualClientImpl struct {
	accrualAddress string
	client         http.Client
}

var _ AccrualClient = (*AccrualClientImpl)(nil)

func NewAccrualClient(accrualAddress string) *AccrualClientImpl {
	return &AccrualClientImpl{accrualAddress: accrualAddress}
}

func (c *AccrualClientImpl) GetAccrual(orderNumber string) (AccrualOrder, error) {
	var result AccrualOrder
	response, err := c.client.Get(c.accrualAddress + getAccrualPath + orderNumber)
	if err != nil {
		return result, fmt.Errorf("get accrual for order %s error: %w", orderNumber, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, fmt.Errorf("read body from accrual API error: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("got status %v from accrual API", response.StatusCode)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("parse body from accrual API error: %w", err)
	}
	return result, nil
}
