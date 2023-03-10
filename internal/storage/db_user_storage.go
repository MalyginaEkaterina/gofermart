package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/gofermart/internal"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DBUserStorage struct {
	db                     *sql.DB
	insertUser             *sql.Stmt
	insertFirstTransaction *sql.Stmt
	getUserByLogin         *sql.Stmt
}

var _ UserStorage = (*DBUserStorage)(nil)

func NewDBUserStorage(db *sql.DB) (*DBUserStorage, error) {
	stmtInsertUser, err := db.Prepare("INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id")
	if err != nil {
		return nil, err
	}
	stmtInsertFirstTransaction, err := db.Prepare("INSERT INTO transactions (id, order_number, user_id, sum, balance, withdrawals) VALUES (0, null, $1, 0, 0, 0)")
	if err != nil {
		return nil, err
	}
	stmtGetUserByLogin, err := db.Prepare("SELECT id, password from users WHERE login = $1")
	if err != nil {
		return nil, err
	}

	return &DBUserStorage{
		db:                     db,
		insertUser:             stmtInsertUser,
		insertFirstTransaction: stmtInsertFirstTransaction,
		getUserByLogin:         stmtGetUserByLogin,
	}, nil
}

func (d *DBUserStorage) Close() {
	d.insertUser.Close()
	d.getUserByLogin.Close()
}

func (d *DBUserStorage) AddUser(ctx context.Context, login string, hashedPass string) (internal.UserID, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction error: %w", err)
	}
	defer tx.Rollback()
	txInsertUserStmt := tx.StmtContext(ctx, d.insertUser)
	row := txInsertUserStmt.QueryRowContext(ctx, login, hashedPass)
	var id internal.UserID
	err = row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrAlreadyExists
	} else if err != nil {
		return 0, fmt.Errorf("insert user error: %w", err)
	}
	txInsertFirstTrStmt := tx.StmtContext(ctx, d.insertFirstTransaction)
	_, err = txInsertFirstTrStmt.ExecContext(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("insert first user transaction error: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("commit error: %w", err)
	}
	return id, nil
}

func (d *DBUserStorage) GetUser(ctx context.Context, login string) (internal.UserID, string, error) {
	row := d.getUserByLogin.QueryRowContext(ctx, login)
	var id internal.UserID
	var pass string
	err := row.Scan(&id, &pass)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNotFound
	} else if err != nil {
		return 0, "", fmt.Errorf("get user error: %w", err)
	}
	return id, pass, nil
}
