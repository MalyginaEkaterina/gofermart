package handlers

import (
	"errors"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/service"
	"io"
	"log"
	"net/http"
)

type OrderHandler struct {
	orderService service.OrderService
}

func (o *OrderHandler) UploadOrder(writer http.ResponseWriter, req *http.Request) {
	userID := GetUserIDFromContext(req.Context())
	orderNum, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if len(orderNum) == 0 {
		http.Error(writer, "Order number is required", http.StatusBadRequest)
		return
	}
	err = o.orderService.UploadOrder(req.Context(), userID, internal.OrderNumber(orderNum))
	if errors.Is(err, service.ErrExistingOrderForUser) {
		writer.WriteHeader(http.StatusOK)
		return
	} else if errors.Is(err, service.ErrExistingOrderForOtherUser) {
		http.Error(writer, err.Error(), http.StatusConflict)
		return
	} else if errors.Is(err, service.ErrIncorrectOrderNumber) {
		http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		log.Println("Upload order error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusAccepted)
}

func (o *OrderHandler) GetOrders(writer http.ResponseWriter, req *http.Request) {
	userID := GetUserIDFromContext(req.Context())
	orders, err := o.orderService.GetOrders(req.Context(), userID)
	if err != nil {
		log.Println("Get orders error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	marshalResponse(writer, http.StatusOK, orders)
}

func (o *OrderHandler) GetBalance(writer http.ResponseWriter, req *http.Request) {
	userID := GetUserIDFromContext(req.Context())
	balance, err := o.orderService.GetBalance(req.Context(), userID)
	if err != nil {
		log.Println("Get balance error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	marshalResponse(writer, http.StatusOK, balance)
}

func (o *OrderHandler) Withdraw(writer http.ResponseWriter, req *http.Request) {
	var withdrawReq internal.WithdrawReq
	if !unmarshalRequest(writer, req, &withdrawReq) {
		return
	}
	if !o.ValidateWithdrawReq(writer, withdrawReq) {
		return
	}
	withdrawReq.UserID = GetUserIDFromContext(req.Context())
	err := o.orderService.Withdraw(req.Context(), withdrawReq)
	if errors.Is(err, service.ErrIncorrectOrderNumber) {
		http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
		return
	} else if errors.Is(err, service.ErrInsufficientFunds) {
		http.Error(writer, err.Error(), http.StatusPaymentRequired)
		return
	} else if err != nil {
		log.Println("Withdraw error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (o *OrderHandler) GetWithdrawals(writer http.ResponseWriter, req *http.Request) {
	userID := GetUserIDFromContext(req.Context())
	withdrawals, err := o.orderService.GetWithdrawals(req.Context(), userID)
	if err != nil {
		log.Println("Get withdrawals error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	marshalResponse(writer, http.StatusOK, withdrawals)
}

func (o *OrderHandler) ValidateWithdrawReq(writer http.ResponseWriter, data internal.WithdrawReq) bool {
	if data.Number == "" {
		http.Error(writer, "Order is required", http.StatusBadRequest)
		return false
	}
	if data.Sum == 0 {
		http.Error(writer, "Sum is required", http.StatusBadRequest)
		return false
	}
	return true
}
