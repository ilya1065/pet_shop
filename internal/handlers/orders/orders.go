package orders

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage/postgres"
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
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.OrderWithItems, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
	PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (orderID int, err error)
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

func writeStorageError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	if errors.Is(err, postgres.ErrNotFound) {
		status = http.StatusNotFound
		message = "not found"
	} else if errors.Is(err, postgres.ErrInvalidInput) {
		status = http.StatusBadRequest
		message = "invalid input"
	}
	w.WriteHeader(status)
	render.JSON(w, r, map[string]string{"error": message})
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
		writeStorageError(w, r, err)
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
		w.WriteHeader(http.StatusBadRequest)
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
		writeStorageError(w, r, err)
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
	if err != nil || id <= 0 {
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
		writeStorageError(w, r, err)
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
	log.Info("getting order and items successful")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, orders)

}

func isEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return email == addr.Address
}

type placeOrderDTO struct {
	Email string             `json:"email"`
	Items []models.OrderItem `json:"items"`
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.PlaceOrder"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))
	log.Info("start place order")
	var req placeOrderDTO
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"message": "error decode json",
		})
		return
	}
	if !isEmail(req.Email) || len(req.Items) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid email or empty items"})
		return
	}
	for _, item := range req.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid order item"})
			return
		}
	}
	orderID, err := h.storage.PlaceOrder(r.Context(), req.Email, req.Items)
	if err != nil {
		log.Error("error place order", slog.Any("error", err))
		writeStorageError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]any{
		"order_id": orderID,
	})

}
