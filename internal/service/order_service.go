package service

import (
	"context"
	"errors"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
)

var (
	ErrExistingOrderForUser      = errors.New("order has already been loaded by this user")
	ErrExistingOrderForOtherUser = errors.New("order has already been loaded by other user")
	ErrIncorrectOrderNumber      = errors.New("incorrect order number")
	ErrInsufficientFunds         = errors.New("there are not enough funds on the account")
)

type OrderService interface {
	UploadOrder(ctx context.Context, userID internal.UserID, number internal.OrderNumber) error
	GetOrders(ctx context.Context, userID internal.UserID) ([]internal.Order, error)
	GetBalance(ctx context.Context, userID internal.UserID) (internal.Balance, error)
	GetWithdrawals(ctx context.Context, userID internal.UserID) ([]internal.Withdrawal, error)
	Withdraw(ctx context.Context, req internal.WithdrawReq) error
}

type OrderServiceImpl struct {
	Store storage.OrderStorage
}

func (o *OrderServiceImpl) UploadOrder(ctx context.Context, userID internal.UserID, number internal.OrderNumber) error {
	if !CheckNumberByLuhn(string(number)) {
		return ErrIncorrectOrderNumber
	}
	err := o.Store.AddOrder(ctx, userID, number)
	if errors.Is(err, storage.ErrAlreadyExists) {
		orderUser, err := o.Store.GetOrderUser(ctx, number)
		if err != nil {
			return err
		}
		if orderUser == userID {
			return ErrExistingOrderForUser
		} else {
			return ErrExistingOrderForOtherUser
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (o *OrderServiceImpl) GetOrders(ctx context.Context, userID internal.UserID) ([]internal.Order, error) {
	return o.Store.GetOrders(ctx, userID)
}

func (o *OrderServiceImpl) GetBalance(ctx context.Context, userID internal.UserID) (internal.Balance, error) {
	return o.Store.GetBalance(ctx, userID)
}

func (o *OrderServiceImpl) GetWithdrawals(ctx context.Context, userID internal.UserID) ([]internal.Withdrawal, error) {
	return o.Store.GetWithdrawals(ctx, userID)
}

func (o *OrderServiceImpl) Withdraw(ctx context.Context, withdraw internal.WithdrawReq) error {
	if !CheckNumberByLuhn(string(withdraw.Number)) {
		return ErrIncorrectOrderNumber
	}
	err := o.Store.Withdraw(ctx, withdraw)
	if errors.Is(err, storage.ErrNotFound) {
		return ErrInsufficientFunds
	}
	return err
}

var _ OrderService = (*OrderServiceImpl)(nil)
