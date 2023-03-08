package handlers

import (
	"encoding/json"
	"github.com/MalyginaEkaterina/gofermart/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
)

func NewRouter(authService service.AuthService, orderService service.OrderService) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gzipHandle)

	authHandler := &AuthHandler{authService: authService}
	orderHandler := &OrderHandler{orderService: orderService}

	r.Post("/api/user/register", authHandler.RegisterUser)
	r.Post("/api/user/login", authHandler.AuthUser)

	authRequiredGroup := r.Group(nil)
	authRequiredGroup.Use(authHandler.Auth)
	authRequiredGroup.Post("/api/user/orders", orderHandler.UploadOrder)
	authRequiredGroup.Get("/api/user/orders", orderHandler.GetOrders)
	authRequiredGroup.Get("/api/user/balance", orderHandler.GetBalance)
	authRequiredGroup.Post("/api/user/balance/withdraw", orderHandler.Withdraw)
	authRequiredGroup.Get("/api/user/withdrawals", orderHandler.GetWithdrawals)

	r.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Wrong request", http.StatusBadRequest)
	})

	r.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Method not allowed", http.StatusBadRequest)
	})
	return r
}

func unmarshalRequest(writer http.ResponseWriter, req *http.Request, v any) bool {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return false
	}
	if len(body) == 0 {
		http.Error(writer, "Request body is required", http.StatusBadRequest)
		return false
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		http.Error(writer, "Failed to parse request body", http.StatusBadRequest)
		return false
	}
	return true
}

func marshalResponse(writer http.ResponseWriter, status int, response any) {
	respJSON, err := json.Marshal(response)
	if err != nil {
		log.Println("Error while serializing response", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(status)
	writer.Write(respJSON)
}
