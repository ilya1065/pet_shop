package orders

import (
	"context"
	"go-pet-shop/internal/models"
	"log/slog"
)

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error) // Возвращает ID созданного заказа
	AddOrderItem(ctx context.Context, orderItem models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (*models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
}

type handler struct {
	log     *slog.Logger
	storage Orders
}

func New(log *slog.Logger, storage Orders) *handler {
	return &handler{
		log:     log,
		storage: storage,
	}
}
