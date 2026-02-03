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
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method is not post"})
		return
	}

	var req dto.CreateOrderRequest
	if err := decodeJSONStrict(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if req.UserID <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "user_id must be positive"})
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "items must not be empty"})
		return
	}

	items := make([]*ordersv1.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.ProductID <= 0 {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "product_id must be positive"})
			return
		}
		if it.Quantity <= 0 {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "quantity must be positive"})
			return
		}
		if it.Price < 0 {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "price must be non-negative"})
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
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.CreateOrderResponse{OrderID: resp.OrderId})
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
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method is not get"})
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "id must be positive"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), g.RequestTimeout)
	defer cancel()

	var req ordersv1.GetOrderRequest
	req.OrderId = id

	resp, err := g.Orders.GetOrder(ctx, &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ProtoGetToDTO(resp))
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
