package handlers

import (
	"context"
	"errors"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/service"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
	"log"
	"net/http"
)

const (
	userIDKey  = authContextKey("userID")
	authHeader = "Authorization"
)

type authContextKey string

type AuthHandler struct {
	authService service.AuthService
}

func (a *AuthHandler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		userID, err := a.authService.CheckToken(req.Header.Get(authHeader))
		if errors.Is(err, service.ErrUnauthorized) {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Println("Check token error: ", err)
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(req.Context(), userIDKey, userID)
		next.ServeHTTP(writer, req.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) internal.UserID {
	return ctx.Value(userIDKey).(internal.UserID)
}

type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a *AuthHandler) RegisterUser(writer http.ResponseWriter, req *http.Request) {
	var registerReq AuthData
	if !unmarshalRequest(writer, req, &registerReq) {
		return
	}
	if !a.ValidateAuthData(writer, registerReq) {
		return
	}
	token, err := a.authService.RegisterUser(req.Context(), registerReq.Login, registerReq.Password)
	if errors.Is(err, storage.ErrAlreadyExists) {
		http.Error(writer, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		log.Println("Register user error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.Header().Set(authHeader, string(token))
	writer.WriteHeader(http.StatusOK)
}

func (a *AuthHandler) AuthUser(writer http.ResponseWriter, req *http.Request) {
	var authReq AuthData
	if !unmarshalRequest(writer, req, &authReq) {
		return
	}
	if !a.ValidateAuthData(writer, authReq) {
		return
	}
	token, err := a.authService.AuthUser(req.Context(), authReq.Login, authReq.Password)
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, service.ErrIncorrectPassword) {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Println("Authentication user error: ", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.Header().Set(authHeader, string(token))
	writer.WriteHeader(http.StatusOK)
}

func (a *AuthHandler) ValidateAuthData(writer http.ResponseWriter, data AuthData) bool {
	if data.Login == "" {
		http.Error(writer, "Login is required", http.StatusBadRequest)
		return false
	}
	if data.Password == "" {
		http.Error(writer, "Password is required", http.StatusBadRequest)
		return false
	}
	return true
}
