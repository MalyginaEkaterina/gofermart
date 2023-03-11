package handlers

import (
	"bytes"
	"context"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/service"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	type want struct {
		statusCode int
		isToken    bool
	}
	tests := []struct {
		name    string
		request string
		store   storage.UserStorage
		want    want
	}{
		{
			name:    "Positive test",
			request: "{\"login\": \"login\",\"password\": \"password\"}",
			store:   &mockUserStorage{userID: 1, loginPass: make(map[string]string)},
			want:    want{statusCode: 200, isToken: true},
		},
		{
			name:    "Negative test with empty body",
			request: "",
			store:   &mockUserStorage{loginPass: make(map[string]string)},
			want:    want{statusCode: 400},
		},
		{
			name:    "Negative test with empty password",
			request: "{\"login\": \"login\"}",
			store:   &mockUserStorage{loginPass: make(map[string]string)},
			want:    want{statusCode: 400},
		},
		{
			name:    "Negative test with used login",
			request: "{\"login\": \"login\",\"password\": \"password\"}",
			store:   &mockUserStorage{addUserErr: storage.ErrAlreadyExists, loginPass: make(map[string]string)},
			want:    want{statusCode: 409},
		},
	}
	orderService := &service.OrderServiceImpl{Store: &mockOrderStorage{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &service.AuthServiceImpl{Store: tt.store, SecretKey: []byte("my secret key")}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			request := httptest.NewRequest(http.MethodPost, ts.URL+"/api/user/register", bytes.NewBufferString(tt.request))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
			if tt.want.isToken {
				assert.Greater(t, len(resp.Header().Get(authHeader)), 0)
			}
		})
	}
}

func TestAuthUser(t *testing.T) {
	type want struct {
		statusCode int
		isToken    bool
	}
	tests := []struct {
		name       string
		preRequest string
		request    string
		store      storage.UserStorage
		want       want
	}{
		{
			name:       "Positive test",
			preRequest: "{\"login\": \"login\",\"password\": \"password\"}",
			request:    "{\"login\": \"login\",\"password\": \"password\"}",
			store:      &mockUserStorage{userID: 1, loginPass: make(map[string]string)},
			want:       want{statusCode: 200, isToken: true},
		},
		{
			name:    "Negative test with empty body",
			request: "",
			store:   &mockUserStorage{loginPass: make(map[string]string)},
			want:    want{statusCode: 400},
		},
		{
			name:    "Negative test with empty password",
			request: "{\"login\": \"login\"}",
			store:   &mockUserStorage{loginPass: make(map[string]string)},
			want:    want{statusCode: 400},
		},
		{
			name:       "Negative test with wrong password",
			preRequest: "{\"login\": \"login\",\"password\": \"password\"}",
			request:    "{\"login\": \"login\",\"password\": \"wrong password\"}",
			store:      &mockUserStorage{loginPass: make(map[string]string)},
			want:       want{statusCode: 401},
		},
	}
	orderService := &service.OrderServiceImpl{Store: &mockOrderStorage{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &service.AuthServiceImpl{Store: tt.store, SecretKey: []byte("my secret key")}
			r := NewRouter(authService, orderService)
			ts := httptest.NewServer(r)
			defer ts.Close()

			if tt.preRequest != "" {
				preRequest := httptest.NewRequest(http.MethodPost, ts.URL+"/api/user/register", bytes.NewBufferString(tt.preRequest))
				preResp := httptest.NewRecorder()
				r.ServeHTTP(preResp, preRequest)
			}

			request := httptest.NewRequest(http.MethodPost, ts.URL+"/api/user/login", bytes.NewBufferString(tt.request))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, request)

			assert.Equal(t, tt.want.statusCode, resp.Code)
			if tt.want.isToken {
				assert.Greater(t, len(resp.Header().Get(authHeader)), 0)
			}
		})
	}
}

type mockUserStorage struct {
	userID     internal.UserID
	loginPass  map[string]string
	addUserErr error
	getUserErr error
}

func (m *mockUserStorage) AddUser(_ context.Context, login string, hashedPass string) (internal.UserID, error) {
	m.loginPass[login] = hashedPass
	return m.userID, m.addUserErr
}

func (m *mockUserStorage) GetUser(_ context.Context, login string) (internal.UserID, string, error) {
	return m.userID, m.loginPass[login], m.getUserErr
}

func (m *mockUserStorage) Close() {
}

var _ storage.UserStorage = (*mockUserStorage)(nil)
