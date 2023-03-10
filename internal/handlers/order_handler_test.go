package handlers

import (
	"bytes"
	"context"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/service"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadOrder(t *testing.T) {
	type want struct {
		statusCode int
	}
	authService := &service.AuthServiceImpl{Store: &mockUserStorage{}, SecretKey: []byte("my secret key")}
	token, err := authService.CreateToken(1)
	require.NoError(t, err)
	tests := []struct {
		name    string
		request string
		token   internal.Token
		store   storage.OrderStorage
		want    want
	}{
		{
			name:    "Positive test",
			request: "8788770",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 202},
		},
		{
			name:    "Test with already uploaded order",
			request: "8788770",
			token:   token,
			store: &mockOrderStorage{
				addOrderErr: storage.ErrAlreadyExists,
				userID:      1,
			},
			want: want{statusCode: 200},
		},
		{
			name:    "Test with already uploaded order by other user",
			request: "8788770",
			token:   token,
			store: &mockOrderStorage{
				addOrderErr: storage.ErrAlreadyExists,
				userID:      2,
			},
			want: want{statusCode: 409},
		},
		{
			name:    "Test without token",
			request: "8788770",
			token:   "",
			store:   &mockOrderStorage{},
			want:    want{statusCode: 401},
		},
		{
			name:    "Test with wrong order number",
			request: "8788470",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 422},
		},
		{
			name:    "Test with wrong request",
			request: "",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 400},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService := &service.OrderServiceImpl{Store: tt.store}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			request := httptest.NewRequest(http.MethodPost, ts.URL+"/api/user/orders", bytes.NewBufferString(tt.request))
			request.Header.Set(authHeader, string(tt.token))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
		})
	}
}

func TestGetOrders(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	authService := &service.AuthServiceImpl{Store: &mockUserStorage{}, SecretKey: []byte("my secret key")}
	token, err := authService.CreateToken(1)
	require.NoError(t, err)
	tests := []struct {
		name  string
		token internal.Token
		store storage.OrderStorage
		want  want
	}{
		{
			name:  "Positive test",
			token: token,
			store: &mockOrderStorage{
				orders: []internal.Order{
					{
						Number:     "9278923470",
						Status:     "PROCESSED",
						Accrual:    float(500),
						UploadedAt: "2020-12-10T15:15:45+03:00",
					},
					{
						Number:     "12345678903",
						Status:     "PROCESSING",
						UploadedAt: "2020-12-10T15:12:01+03:00",
					},
				},
			},
			want: want{
				statusCode: 200,
				response:   "[{\"number\":\"9278923470\",\"status\":\"PROCESSED\",\"accrual\":500,\"uploaded_at\":\"2020-12-10T15:15:45+03:00\"},{\"number\":\"12345678903\",\"status\":\"PROCESSING\",\"uploaded_at\":\"2020-12-10T15:12:01+03:00\"}]"},
		},
		{
			name:  "Test with no content",
			token: token,
			store: &mockOrderStorage{orders: []internal.Order{}},
			want:  want{statusCode: 204},
		},
		{
			name:  "Test without token",
			token: "",
			store: &mockOrderStorage{orders: []internal.Order{}},
			want:  want{statusCode: 401},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService := &service.OrderServiceImpl{Store: tt.store}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			request := httptest.NewRequest(http.MethodGet, ts.URL+"/api/user/orders", nil)
			request.Header.Set(authHeader, string(tt.token))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
			if tt.want.response != "" {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(respBody))
			}
		})
	}

}

func TestGetBalance(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	authService := &service.AuthServiceImpl{Store: &mockUserStorage{}, SecretKey: []byte("my secret key")}
	token, err := authService.CreateToken(1)
	require.NoError(t, err)
	tests := []struct {
		name  string
		token internal.Token
		store storage.OrderStorage
		want  want
	}{
		{
			name:  "Positive test",
			token: token,
			store: &mockOrderStorage{
				balance: internal.Balance{
					Current:   500.5,
					Withdrawn: 42,
				},
			},
			want: want{
				statusCode: 200,
				response:   "{\"current\":500.5,\"withdrawn\":42}"},
		},
		{
			name:  "Test without token",
			token: "",
			store: &mockOrderStorage{},
			want:  want{statusCode: 401},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService := &service.OrderServiceImpl{Store: tt.store}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			request := httptest.NewRequest(http.MethodGet, ts.URL+"/api/user/balance", nil)
			request.Header.Set(authHeader, string(tt.token))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
			if tt.want.response != "" {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(respBody))
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	type want struct {
		statusCode int
	}
	authService := &service.AuthServiceImpl{Store: &mockUserStorage{}, SecretKey: []byte("my secret key")}
	token, err := authService.CreateToken(1)
	require.NoError(t, err)
	tests := []struct {
		name    string
		request string
		token   internal.Token
		store   storage.OrderStorage
		want    want
	}{
		{
			name:    "Positive test",
			request: "{\"order\":\"2377225624\",\"sum\":751}",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 200},
		},
		{
			name:    "Test with wrong request 1",
			request: "{\"sum\":751}",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 400},
		},
		{
			name:    "Test with wrong request 2",
			request: "{\"order\":\"2377225624\"}",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 400},
		},
		{
			name:    "Test without token",
			request: "{\"order\":\"2377225624\",\"sum\":751}",
			token:   "",
			store:   &mockOrderStorage{},
			want:    want{statusCode: 401},
		},
		{
			name:    "Test with wrong order number",
			request: "{\"order\":\"235624\",\"sum\":751}",
			token:   token,
			store:   &mockOrderStorage{},
			want:    want{statusCode: 422},
		},
		{
			name:    "Test with insufficient balance",
			request: "{\"order\":\"2377225624\",\"sum\":751}",
			token:   token,
			store:   &mockOrderStorage{withdrawErr: storage.ErrNotFound},
			want:    want{statusCode: 402},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService := &service.OrderServiceImpl{Store: tt.store}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			request := httptest.NewRequest(http.MethodPost, ts.URL+"/api/user/balance/withdraw", bytes.NewBufferString(tt.request))
			request.Header.Set(authHeader, string(tt.token))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	authService := &service.AuthServiceImpl{Store: &mockUserStorage{}, SecretKey: []byte("my secret key")}
	token, err := authService.CreateToken(1)
	require.NoError(t, err)
	tests := []struct {
		name  string
		token internal.Token
		store storage.OrderStorage
		want  want
	}{
		{
			name:  "Positive test",
			token: token,
			store: &mockOrderStorage{
				withdrawals: []internal.Withdrawal{
					{
						Number:      "2377225624",
						Sum:         float64(500),
						ProcessedAt: "2020-12-09T16:09:57+03:00",
					},
				},
			},
			want: want{
				statusCode: 200,
				response:   "[{\"order\":\"2377225624\",\"sum\":500,\"processed_at\":\"2020-12-09T16:09:57+03:00\"}]",
			},
		},
		{
			name:  "Test without token",
			token: "",
			store: &mockOrderStorage{},
			want:  want{statusCode: 401},
		},
		{
			name:  "Test with no content",
			token: token,
			store: &mockOrderStorage{withdrawals: []internal.Withdrawal{}},
			want:  want{statusCode: 204},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService := &service.OrderServiceImpl{Store: tt.store}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			request := httptest.NewRequest(http.MethodGet, ts.URL+"/api/user/withdrawals", nil)
			request.Header.Set(authHeader, string(tt.token))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
			if tt.want.response != "" {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(respBody))
			}
		})
	}
}

type mockOrderStorage struct {
	userID            internal.UserID
	addOrderErr       error
	orders            []internal.Order
	getOrdersErr      error
	balance           internal.Balance
	withdrawErr       error
	withdrawals       []internal.Withdrawal
	getWithdrawalsErr error
}

func (m *mockOrderStorage) AddOrder(_ context.Context, _ internal.UserID, _ internal.OrderNumber) error {
	return m.addOrderErr
}

func (m *mockOrderStorage) GetOrderUser(_ context.Context, _ internal.OrderNumber) (internal.UserID, error) {
	return m.userID, nil
}

func (m *mockOrderStorage) GetOrdersByUser(_ context.Context, _ internal.UserID) ([]internal.Order, error) {
	return m.orders, m.getOrdersErr
}

func (m *mockOrderStorage) GetNotProcessedOrders(_ context.Context) ([]internal.ProcessingOrder, error) {
	return nil, nil
}

func (m *mockOrderStorage) UpdateOrderStatus(_ context.Context, _ internal.ProcessingOrder) error {
	return nil
}

func (m *mockOrderStorage) UpdateOrderAccrual(_ context.Context, _ internal.ProcessingOrder) error {
	return nil
}

func (m *mockOrderStorage) GetBalance(_ context.Context, _ internal.UserID) (internal.Balance, error) {
	return m.balance, nil
}

func (m *mockOrderStorage) GetWithdrawals(_ context.Context, _ internal.UserID) ([]internal.Withdrawal, error) {
	return m.withdrawals, m.getWithdrawalsErr
}

func (m *mockOrderStorage) Withdraw(_ context.Context, _ internal.WithdrawReq) error {
	return m.withdrawErr
}

func (m *mockOrderStorage) Close() {
}

var _ storage.OrderStorage = (*mockOrderStorage)(nil)

func float(v float64) *float64 {
	return &v
}
