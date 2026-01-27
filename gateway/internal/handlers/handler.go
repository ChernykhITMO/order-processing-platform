package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/dto"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

type Gateway struct {
	Orders         ordersv1.OrdersServiceClient
	RequestTimeout time.Duration
}

// HandleOrders godoc
// @Summary Create order
// @Description Create a new order
// @Accept json
// @Produce json
// @Param request body dto.CreateOrderRequest true "Create order"
// @Success 200 {object} dto.CreateOrderResponse
// @Failure 400 {string} string
// @Router /orders [post]
func (g *Gateway) HandleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method is not post", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req dto.CreateOrderRequest
	if err := decodeJSONStrict(w, r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.UserID <= 0 {
		http.Error(w, "user_id must be positive", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items must not be empty", http.StatusBadRequest)
		return
	}

	items := make([]*ordersv1.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.ProductID <= 0 {
			http.Error(w, "product_id must be positive", http.StatusBadRequest)
			return
		}
		if it.Quantity <= 0 {
			http.Error(w, "quantity must be positive", http.StatusBadRequest)
			return
		}
		if it.Price < 0 {
			http.Error(w, "price must be non-negative", http.StatusBadRequest)
			return
		}
		items = append(items, &ordersv1.OrderItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
			Price:     &ordersv1.Money{Money: it.Price},
		})
	}

	ctx, cancel := context.WithTimeout(r.Context(), g.RequestTimeout)
	defer cancel()

	var protoReq ordersv1.CreateOrderRequest
	protoReq.UserId = req.UserID
	protoReq.Items = items

	resp, err := g.Orders.CreateOrder(ctx, &protoReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	var dtoResp dto.CreateOrderResponse
	dtoResp.OrderID = resp.OrderId

	if err := json.NewEncoder(w).Encode(&dtoResp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// HandleOrderById godoc
// @Summary Get order by id
// @Description Returns order details by id
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} dto.GetOrderResponse
// @Failure 400 {string} string
// @Router /orders/{id} [get]
func (g *Gateway) HandleOrderById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method is not get", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id <= 0 {
		http.Error(w, "id must be positive", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), g.RequestTimeout)
	defer cancel()

	var req ordersv1.GetOrderRequest
	req.OrderId = id

	resp, err := g.Orders.GetOrder(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	dtoResp := dto.ProtoGetToDTO(resp)

	if err := json.NewEncoder(w).Encode(&dtoResp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func decodeJSONStrict(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("unexpected trailing data")
	}

	return nil
}
