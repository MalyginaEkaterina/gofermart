package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type UserStorage interface {
	AddUser(ctx context.Context, login string, hashedPass string) (internal.UserID, error)
	GetUser(ctx context.Context, login string) (internal.UserID, string, error)
	Close()
}

type OrderStorage interface {
	AddOrder(ctx context.Context, userID internal.UserID, number internal.OrderNumber) error
	GetOrderUser(ctx context.Context, number internal.OrderNumber) (internal.UserID, error)
	GetOrdersByUser(ctx context.Context, userID internal.UserID) ([]internal.Order, error)
	GetNotProcessedOrders(ctx context.Context) ([]internal.ProcessingOrder, error)
	UpdateOrderStatus(ctx context.Context, order internal.ProcessingOrder) error
	UpdateOrderAccrual(ctx context.Context, order internal.ProcessingOrder) error
	GetBalance(ctx context.Context, userID internal.UserID) (internal.Balance, error)
	GetWithdrawals(ctx context.Context, userID internal.UserID) ([]internal.Withdrawal, error)
	Withdraw(ctx context.Context, withdraw internal.WithdrawReq) error
	Close()
}

func DoMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
