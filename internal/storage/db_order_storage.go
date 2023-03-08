package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"time"
)

type DBOrderStorage struct {
	db                      *sql.DB
	insertOrder             *sql.Stmt
	getOrderUserByNumber    *sql.Stmt
	getOrdersByUser         *sql.Stmt
	getNotProcessedOrders   *sql.Stmt
	updateOrderStatus       *sql.Stmt
	updateOrder             *sql.Stmt
	insertCreditTransaction *sql.Stmt
	getBalance              *sql.Stmt
	insertDebitTransaction  *sql.Stmt
	getWithdrawalsByUser    *sql.Stmt
}

var _ OrderStorage = (*DBOrderStorage)(nil)

func NewDBOrderStorage(db *sql.DB) (*DBOrderStorage, error) {
	stmtInsertOrder, err := db.Prepare("INSERT INTO orders (number, user_id, status) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING number")
	if err != nil {
		return nil, err
	}
	stmtGetOrderUserByNumber, err := db.Prepare("SELECT user_id from orders WHERE number = $1")
	if err != nil {
		return nil, err
	}
	stmtGetOrdersByUser, err := db.Prepare("SELECT number, status, accrual, uploaded_at from orders WHERE user_id = $1 ORDER BY uploaded_at")
	if err != nil {
		return nil, err
	}
	stmtGetNotProcessedOrders, err := db.Prepare("SELECT number, status, accrual, user_id from orders WHERE status not in ($1, $2)")
	if err != nil {
		return nil, err
	}
	stmtUpdateOrderStatus, err := db.Prepare("UPDATE orders SET status = $1 WHERE number = $2")
	if err != nil {
		return nil, err
	}
	stmtUpdateOrder, err := db.Prepare("UPDATE orders SET status = $1, accrual = $2 WHERE number = $3")
	if err != nil {
		return nil, err
	}
	const insertCreditTrSQL = `
		INSERT INTO transactions (id, number, user_id, sum, balance, withdrawals)
		SELECT id+1,
				 $1,
				 $2,
				 $3,
				 balance+$3,
				 withdrawals
		FROM transactions
		WHERE user_id = $2
		ORDER BY id DESC LIMIT 1
	`
	stmtInsertCreditTransaction, err := db.Prepare(insertCreditTrSQL)
	if err != nil {
		return nil, err
	}
	stmtGetBalance, err := db.Prepare("SELECT balance, withdrawals FROM transactions WHERE user_id = $1 ORDER BY id DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	const insertDebitTrSQL = `
		INSERT INTO transactions (id, number, user_id, sum, balance, withdrawals)
		(SELECT id + 1,
			   $1,
			   $2,
			   -1.0 * $3,
			   balance - $3,
			   withdrawals + $3
		FROM
			(SELECT id,
					balance,
					withdrawals
			 FROM transactions
			 WHERE user_id = $2
			 ORDER BY id DESC LIMIT 1) A
		WHERE balance >= $3)
		RETURNING id
	`
	stmtInsertDebitTransaction, err := db.Prepare(insertDebitTrSQL)
	if err != nil {
		return nil, err
	}
	stmtGetWithdrawalsByUser, err := db.Prepare("SELECT number, -1*sum, processed_at FROM transactions WHERE user_id = $1 AND sum < 0 ORDER BY processed_at")
	if err != nil {
		return nil, err
	}
	return &DBOrderStorage{
		db:                      db,
		insertOrder:             stmtInsertOrder,
		getOrderUserByNumber:    stmtGetOrderUserByNumber,
		getOrdersByUser:         stmtGetOrdersByUser,
		getNotProcessedOrders:   stmtGetNotProcessedOrders,
		updateOrderStatus:       stmtUpdateOrderStatus,
		updateOrder:             stmtUpdateOrder,
		insertCreditTransaction: stmtInsertCreditTransaction,
		getBalance:              stmtGetBalance,
		insertDebitTransaction:  stmtInsertDebitTransaction,
		getWithdrawalsByUser:    stmtGetWithdrawalsByUser,
	}, nil
}

func (d *DBOrderStorage) Close() {
	d.insertOrder.Close()
	d.getOrderUserByNumber.Close()
	d.getOrdersByUser.Close()
	d.getNotProcessedOrders.Close()
	d.updateOrderStatus.Close()
	d.updateOrder.Close()
	d.insertCreditTransaction.Close()
	d.getBalance.Close()
	d.insertDebitTransaction.Close()
	d.getWithdrawalsByUser.Close()
}

func (d *DBOrderStorage) AddOrder(ctx context.Context, userID internal.UserID, number internal.OrderNumber) error {
	row := d.insertOrder.QueryRowContext(ctx, number, userID, internal.New)
	var id int
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAlreadyExists
	} else if err != nil {
		return fmt.Errorf("insert order error: %w", err)
	}
	return nil
}

func (d *DBOrderStorage) GetOrderUser(ctx context.Context, number internal.OrderNumber) (internal.UserID, error) {
	row := d.getOrderUserByNumber.QueryRowContext(ctx, number)
	var userID internal.UserID
	err := row.Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, fmt.Errorf("get order user error: %w", err)
	}
	return userID, nil
}

func (d *DBOrderStorage) GetOrders(ctx context.Context, userID internal.UserID) ([]internal.Order, error) {
	rows, err := d.getOrdersByUser.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userOrders []internal.Order

	for rows.Next() {
		order := internal.Order{}
		var uploadTime time.Time
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &uploadTime)
		if err != nil {
			return nil, err
		}
		order.UploadedAt = uploadTime.Format(time.RFC3339)
		userOrders = append(userOrders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return userOrders, nil
}

func (d *DBOrderStorage) GetNotProcessedOrders(ctx context.Context) ([]internal.ProcessingOrder, error) {
	rows, err := d.getNotProcessedOrders.QueryContext(ctx, internal.Processed, internal.Invalid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []internal.ProcessingOrder

	for rows.Next() {
		order := internal.ProcessingOrder{}
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UserID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (d *DBOrderStorage) UpdateOrderStatus(ctx context.Context, order internal.ProcessingOrder) error {
	_, err := d.updateOrderStatus.ExecContext(ctx, order.Status, order.Number)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBOrderStorage) UpdateOrderAccrual(ctx context.Context, order internal.ProcessingOrder) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}
	defer tx.Rollback()
	txUpdateOrderStmt := tx.StmtContext(ctx, d.updateOrder)
	_, err = txUpdateOrderStmt.ExecContext(ctx, order.Status, order.Accrual, order.Number)
	if err != nil {
		return fmt.Errorf("update order error: %w", err)
	}
	txInsertTrStmt := tx.StmtContext(ctx, d.insertCreditTransaction)
	_, err = txInsertTrStmt.ExecContext(ctx, order.Number, order.UserID, order.Accrual)
	if err != nil {
		return fmt.Errorf("insert credit transaction error: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	return nil
}

func (d *DBOrderStorage) GetBalance(ctx context.Context, userID internal.UserID) (internal.Balance, error) {
	row := d.getBalance.QueryRowContext(ctx, userID)
	var balance internal.Balance
	err := row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return balance, fmt.Errorf("get balance error: %w", err)
	}
	return balance, nil
}

func (d *DBOrderStorage) Withdraw(ctx context.Context, withdraw internal.WithdrawReq) error {
	row := d.insertDebitTransaction.QueryRowContext(ctx, withdraw.Number, withdraw.UserID, withdraw.Sum)
	var id int
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("insert debit transaction error: %w", err)
	}
	return nil
}

func (d *DBOrderStorage) GetWithdrawals(ctx context.Context, userID internal.UserID) ([]internal.Withdrawal, error) {
	rows, err := d.getWithdrawalsByUser.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userWithdrawals []internal.Withdrawal

	for rows.Next() {
		withdrawal := internal.Withdrawal{}
		var processedTime time.Time
		err = rows.Scan(&withdrawal.Number, &withdrawal.Sum, &processedTime)
		if err != nil {
			return nil, err
		}
		withdrawal.ProcessedAt = processedTime.Format(time.RFC3339)
		userWithdrawals = append(userWithdrawals, withdrawal)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return userWithdrawals, nil
}
