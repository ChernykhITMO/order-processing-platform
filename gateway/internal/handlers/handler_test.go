package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/dto"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ordersClientMock struct {
	createResp *ordersv1.CreateOrderResponse
	createErr  error
	getResp    *ordersv1.GetOrderResponse
	getErr     error

	lastCreate *ordersv1.CreateOrderRequest
	lastGet    *ordersv1.GetOrderRequest
}

func (m *ordersClientMock) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest, _ ...grpc.CallOption) (*ordersv1.CreateOrderResponse, error) {
	m.lastCreate = req
	return m.createResp, m.createErr
}

func (m *ordersClientMock) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest, _ ...grpc.CallOption) (*ordersv1.GetOrderResponse, error) {
	m.lastGet = req
	return m.getResp, m.getErr
}

func TestHandleOrders_ValidationErrors(t *testing.T) {
	client := &ordersClientMock{}
	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	cases := []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{"wrong method", http.MethodGet, "", http.StatusMethodNotAllowed},
		{"bad json", http.MethodPost, "{", http.StatusBadRequest},
		{"unknown field", http.MethodPost, `{"user_id":1,"items":[],"extra":1}`, http.StatusBadRequest},
		{"invalid user", http.MethodPost, `{"user_id":0,"items":[{"product_id":1,"quantity":1,"price":1}]}`, http.StatusBadRequest},
		{"empty items", http.MethodPost, `{"user_id":1,"items":[]}`, http.StatusBadRequest},
		{"invalid item", http.MethodPost, `{"user_id":1,"items":[{"product_id":0,"quantity":1,"price":1}]}`, http.StatusBadRequest},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/orders", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			gateway.HandleOrders(w, req)
			if w.Code != tt.want {
				data, _ := io.ReadAll(w.Body)
				t.Fatalf("status: got %d, want %d, body: %s", w.Code, tt.want, string(data))
			}
		})
	}
}

func TestHandleOrders_Success(t *testing.T) {
	client := &ordersClientMock{createResp: &ordersv1.CreateOrderResponse{OrderId: 42}}
	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	payload := dto.CreateOrderRequest{
		UserID: 1,
		Items:  []dto.OrderItem{{ProductID: 10, Quantity: 2, Price: 150}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()

	gateway.HandleOrders(w, req)
	if w.Code != http.StatusOK {
		data, _ := io.ReadAll(w.Body)
		t.Fatalf("status: got %d, body: %s", w.Code, string(data))
	}

	if client.lastCreate == nil {
		t.Fatalf("expected CreateOrder to be called")
	}
	if client.lastCreate.UserId != payload.UserID {
		t.Fatalf("user id: got %d, want %d", client.lastCreate.UserId, payload.UserID)
	}
	if len(client.lastCreate.Items) != len(payload.Items) {
		t.Fatalf("items length: got %d, want %d", len(client.lastCreate.Items), len(payload.Items))
	}
}

func TestHandleOrderById_ValidationErrors(t *testing.T) {
	client := &ordersClientMock{}
	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	cases := []struct {
		name string
		url  string
		want int
	}{
		{"wrong method", "/orders/1", http.StatusMethodNotAllowed},
		{"bad id", "/orders/not-number", http.StatusBadRequest},
		{"zero id", "/orders/0", http.StatusBadRequest},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.name == "wrong method" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, tt.url, nil)
			w := httptest.NewRecorder()

			gateway.HandleOrderById(w, req)
			if w.Code != tt.want {
				data, _ := io.ReadAll(w.Body)
				t.Fatalf("status: got %d, want %d, body: %s", w.Code, tt.want, string(data))
			}
		})
	}
}

func TestHandleOrderById_Success(t *testing.T) {
	now := time.Now()
	client := &ordersClientMock{getResp: &ordersv1.GetOrderResponse{Order: &ordersv1.Order{
		OrderId: 1,
		UserId:  2,
		Status:  ordersv1.OrderStatus_new,
		Items: []*ordersv1.OrderItem{{
			ProductId: 10,
			Quantity:  2,
			Price:     &ordersv1.Money{Money: 100},
		}},
		TotalAmount: &ordersv1.Money{Money: 200},
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}}}

	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()

	gateway.HandleOrderById(w, req)
	if w.Code != http.StatusOK {
		data, _ := io.ReadAll(w.Body)
		t.Fatalf("status: got %d, body: %s", w.Code, string(data))
	}

	var resp dto.GetOrderResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Order.OrderID != 1 || resp.Order.UserID != 2 {
		t.Fatalf("unexpected order: %+v", resp.Order)
	}
	if resp.Order.Status != "new" {
		t.Fatalf("status: got %s", resp.Order.Status)
	}
}

func TestHandleOrders_UpstreamError(t *testing.T) {
	client := &ordersClientMock{createErr: errors.New("grpc error")}
	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	payload := dto.CreateOrderRequest{
		UserID: 1,
		Items:  []dto.OrderItem{{ProductID: 10, Quantity: 1, Price: 10}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()

	gateway.HandleOrders(w, req)
	if w.Code != http.StatusBadRequest {
		data, _ := io.ReadAll(w.Body)
		t.Fatalf("status: got %d, body: %s", w.Code, string(data))
	}
}

func TestHandleOrderById_UpstreamError(t *testing.T) {
	client := &ordersClientMock{getErr: errors.New("grpc error")}
	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()

	gateway.HandleOrderById(w, req)
	if w.Code != http.StatusBadRequest {
		data, _ := io.ReadAll(w.Body)
		t.Fatalf("status: got %d, body: %s", w.Code, string(data))
	}

	if client.lastGet == nil || client.lastGet.OrderId != 1 {
		t.Fatalf("expected GetOrder to be called with id=1")
	}
}

func TestHandleOrderById_ParseId(t *testing.T) {
	client := &ordersClientMock{getResp: &ordersv1.GetOrderResponse{Order: &ordersv1.Order{OrderId: 7}}}
	gateway := &Gateway{Orders: client, RequestTimeout: time.Second}

	req := httptest.NewRequest(http.MethodGet, "/orders/"+strconv.FormatInt(7, 10), nil)
	w := httptest.NewRecorder()

	gateway.HandleOrderById(w, req)
	if w.Code != http.StatusOK {
		data, _ := io.ReadAll(w.Body)
		t.Fatalf("status: got %d, body: %s", w.Code, string(data))
	}
}
