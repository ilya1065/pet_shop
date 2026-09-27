package user

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage/postgres"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Users interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Hendler struct {
	log     *slog.Logger
	storage Users
}

func New(log *slog.Logger, users Users) Hendler {
	return Hendler{
		log:     log,
		storage: users,
	}
}

func setupLogger(fn string, ctx context.Context, logger *slog.Logger) *slog.Logger {
	return logger.With(
		slog.String("fn", fn),
		slog.String("request id", middleware.GetReqID(ctx)),
	)
}

func validateUser(name, email string) map[string]string {
	validationErrors := make(map[string]string)
	if name == "" {
		validationErrors["name"] = "name is empty"
	}
	if len(name) > 50 {
		validationErrors["name"] = "the name is too long > 50"
	}
	if len(name) < 2 {
		validationErrors["name"] = "the name is too short < 2"
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		validationErrors["email"] = "invalid email"
	}
	if len(validationErrors) == 0 {
		return nil
	}
	return validationErrors
}

func (h Hendler) CreateUser(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.CreateUser"
	log := setupLogger(fn, r.Context(), h.log)
	log.Info("start creating user", slog.String("url", r.URL.String()))

	var user models.User
	if err := render.DecodeJSON(r.Body, &user); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "bad request", "message": "invalid JSON payload"})
		return
	}

	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	if validationErrors := validateUser(user.Name, user.Email); validationErrors != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, validationErrors)
		return
	}

	id, err := h.storage.CreateUser(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]any{"id": id, "user": user})
}

func (h Hendler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "bad request"})
		return
	}

	user, err := h.storage.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]string{"error": "not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "internal server error"})
		return
	}

	render.JSON(w, r, user)
}

func (h Hendler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.storage.GetAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "internal server error"})
		return
	}

	render.JSON(w, r, users)
}
