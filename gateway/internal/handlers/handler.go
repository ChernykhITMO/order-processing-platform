package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/api"
	"github.com/ChernykhITMO/order-processing-platform/protos/gen/ordersv1"
)

type Gateway struct {
	Orders ordersv1.OrdersServiceClient
}

/*
func (g *HTTPGateway) RunServer() error {

	srvChan := make(chan error)

	go func() {
		srvChan <- g.srv.ListenAndServe()
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		// TODO: научиться пользовать логированием
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return g.srv.Shutdown(ctx)
	case err := <-srvChan:
		return err
	}
} */

func (g *Gateway) HandleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method is not post", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req api.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	items := make([]*ordersv1.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &ordersv1.OrderItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
			Price: &ordersv1.Money{
				Units: it.Price.Units,
				Nanos: it.Price.Nanos,
			},
		})
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	resp, err := g.Orders.CreateOrder(ctx, &ordersv1.CreateOrderRequest{
		UserId: req.UserID,
		Items:  items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		// TODO
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (g *Gateway) HandleOrderById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method is not get", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	resp, err := g.Orders.GetOrder(ctx, &ordersv1.GetOrderRequest{OrderId: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
