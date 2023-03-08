package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUnauthorized      = errors.New("unauthorized")
)

const (
	tokenExpiry = 72 * time.Hour
)

type AuthService interface {
	RegisterUser(ctx context.Context, login string, pass string) (internal.Token, error)
	AuthUser(ctx context.Context, login string, pass string) (internal.Token, error)
	CheckToken(token string) (internal.UserID, error)
}

type AuthServiceImpl struct {
	Store     storage.UserStorage
	SecretKey []byte
}

func (a *AuthServiceImpl) RegisterUser(ctx context.Context, login string, pass string) (internal.Token, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("create password hash error: %w", err)
	}
	userID, err := a.Store.AddUser(ctx, login, hex.EncodeToString(hashedPass))
	if err != nil {
		return "", err
	}
	token, err := a.CreateToken(userID)
	if err != nil {
		return "", fmt.Errorf("create token error: %w", err)
	}
	return token, nil
}

func (a *AuthServiceImpl) AuthUser(ctx context.Context, login string, pass string) (internal.Token, error) {
	userID, hashedPass, err := a.Store.GetUser(ctx, login)
	if err != nil {
		return "", err
	}
	hash, err := hex.DecodeString(hashedPass)
	if err != nil {
		return "", fmt.Errorf("decode hashed password error: %w", err)
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(pass))
	if err != nil {
		return "", ErrIncorrectPassword
	}
	token, err := a.CreateToken(userID)
	if err != nil {
		return "", fmt.Errorf("create token error: %w", err)
	}
	return token, nil
}

func (a *AuthServiceImpl) CreateToken(id internal.UserID) (internal.Token, error) {
	data := binary.BigEndian.AppendUint32(nil, uint32(id))
	expTime := time.Now().Add(tokenExpiry).Unix()
	data = binary.BigEndian.AppendUint64(data, uint64(expTime))
	h := hmac.New(sha256.New, a.SecretKey)
	h.Write(data)
	sign := h.Sum(nil)
	data = append(data, sign...)
	return internal.Token(hex.EncodeToString(data)), nil
}

func (a *AuthServiceImpl) CheckToken(token string) (internal.UserID, error) {
	data, err := hex.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("decode token error: %w", err)
	}
	if len(data) < 12 {
		return 0, ErrUnauthorized
	}
	userID := binary.BigEndian.Uint32(data[:4])
	expTime := binary.BigEndian.Uint64(data[4:12])
	if time.Now().Unix() > int64(expTime) {
		return 0, ErrUnauthorized
	}
	h := hmac.New(sha256.New, a.SecretKey)
	h.Write(data[:12])
	sign := h.Sum(nil)
	if !hmac.Equal(sign, data[12:]) {
		return 0, ErrUnauthorized
	}
	return internal.UserID(userID), nil
}

var _ AuthService = (*AuthServiceImpl)(nil)
