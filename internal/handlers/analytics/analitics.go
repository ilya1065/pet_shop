package analytics

import (
	"context"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/mail"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Analytics interface {
	GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error)
	GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error)
}

type Hendler struct {
	log     *slog.Logger
	storage Analytics
}

func New(log *slog.Logger, storage Analytics) Hendler {
	return Hendler{
		log:     log,
		storage: storage,
	}
}
func setupLogger(logger *slog.Logger, fn, reqID string) *slog.Logger {
	return logger.With(slog.String("fn", fn), slog.String("request id", reqID))
}

func (h *Hendler) GetUserOrderHistory(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.CreateOrder"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))
	log.Info("start Get user order history")

	email := r.URL.Query().Get("email")
	if !isEmail(email) {
		log.Error("invalid email")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad requst",
			"message": "invalid email",
		})
		return
	}
	irderDetail, err := h.storage.GetUserOrderHistory(r.Context(), email)
	if err != nil {
		log.Error("filed get user order history", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"message": "filed get user order history",
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, irderDetail)

}
func isEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return email == addr.Address
}

func (h *Hendler) GetPopularProducts(w http.ResponseWriter, r *http.Request) {
	const fn = "hendlers.orders.CreateOrder"
	log := setupLogger(h.log, fn, middleware.GetReqID(r.Context()))
	log.Info("start get popular products")

	popularProduct, err := h.storage.GetPopularProducts(r.Context())
	if err != nil {
		log.Error("error getting popular product", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"message": "error getting popular product",
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, popularProduct)
}
