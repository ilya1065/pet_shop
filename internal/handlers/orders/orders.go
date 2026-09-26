package orders

import (
	"context"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error) // Возвращает ID созданного заказа
	AddOrderItem(ctx context.Context, orderItem models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (*models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
}

type Handler struct {
	log     *slog.Logger
	storage Orders
}

func New(log *slog.Logger, storage Orders) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func setupLogger(logger *slog.Logger, fn, reqID string) *slog.Logger {
	return logger.With(slog.String("fn", fn), slog.String("request id", reqID))
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.CreateOrder"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))

	log.Info("Start creating order")
	var order models.Order
	err := render.DecodeJSON(r.Body, &order)
	if err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{})
		return
	}
	if order.CustomerID <= 0 {
		slog.Error("customerID <=0")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error": "customerID <= 0",
		})
		return
	}

	id, err := h.storage.CreateOrder(r.Context(), order)
	if err != nil {
		log.Error("error created order", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"message": "failed creating orders",
		})
		return
	}
	log.Info("creating order successful ")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]int{
		"id": id,
	})

}

func (h *Handler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.CreateOrder"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))
	log.Info("start add order item to item")

	idstr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		log.Error("invalid id", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massage": "invalid id",
		})
		return
	}
	if id <= 0 {
		log.Error("id <= 0")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massege": "invalid id",
		})
		return
	}

	var item models.OrderItem
	err = render.DecodeJSON(r.Body, &item)
	if err != nil {
		log.Error("invalid json", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massage": "invalid json",
		})
		return
	}

	if item.Quantity <= 0 {
		log.Error("quantity <= 0")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massage": " invalid quantity",
		})
		return
	}

	if item.ProductID <= 0 {
		log.Error("productID <= 0")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massage": "invalid product id",
		})
		return
	}
	item.OrderID = id
	err = h.storage.AddOrderItem(r.Context(), item)
	if err != nil {
		log.Error("error add order item", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"massege": "filed add order item",
		})
		return
	}
	log.Info("add order item successful")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]string{
		"massege": "add order item successful",
	})

}
func (h *Handler) GetTheOrderDetails(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.GetTheOrderDetails"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))
	log.Info("start getting details order")

	strId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(strId)
	if err != nil {
		slog.Error("error conversion strID to id ")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"message": "invalid id",
		})
		return
	}
	order, err := h.storage.GetOrderByID(r.Context(), id)
	if err != nil {
		log.Error("error getting order by id ", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"massage": "error getting order details",
		})
		return
	}
	orderItems, err := h.storage.GetOrderItemsByOrderID(r.Context(), order.ID)
	if err != nil {
		log.Error("error getting order items", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"massage": "failed getting order items",
		})
		return
	}
	log.Info("getting order and order details successful")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]any{
		"status":      "getting order and order details successful",
		"order":       order,
		"order items": orderItems,
	})

}
func (h *Handler) GetOrderUserByEmail(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.GetOrderUserByEmail"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))
	log.Info("start getting order user by email")

	email := r.URL.Query().Get("email")

	if !isEmail(email) {
		log.Error("invalid email")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massage": "invalid email",
		})
		return
	}
	orders, err := h.storage.GetOrdersByUserEmail(r.Context(), email)

	if err != nil {
		log.Error("error getting orders", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"message": "failed getting orders",
		})
		return
	}
	var fullOrders []OrderWithItems

	for _, order := range orders {

		items, err := h.storage.GetOrderItemsByOrderID(r.Context(), order.ID)
		if err != nil {
			log.Error("error getting items")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"error":   "internal server error",
				"massage": "failed getting items",
			})
			return
		}
		fullOrders = append(fullOrders, OrderWithItems{
			Order: order,
			Items: items,
		})
	}

	log.Info("getting order and items successful")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, fullOrders)

}

type OrderWithItems struct {
	Order models.Order       `json:"order"`
	Items []models.OrderItem `json:"items"`
}

func isEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return email == addr.Address
}
