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

func New(log *slog.Logger, Users Users) Hendler {
	return Hendler{
		log:     log,
		storage: Users,
	}
}
func setupLogger(fn string, ctx context.Context, logger *slog.Logger) *slog.Logger {
	return logger.With(
		slog.String("fn", fn),
		slog.String("request id", middleware.GetReqID(ctx)),
	)
}

func validateUser(name, email string) map[string]string {
	errors := make(map[string]string)
	if name == "" {
		errors["name"] = "name is empty"
	}
	if len(name) > 50 {
		errors["name"] = "the name is too long > 50"
	}
	if len(name) < 2 {
		errors["name"] = "the name is too short < 2"
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		errors["email"] = "invalid email"
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

func (h Hendler) CreateUser(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.CreateUser"
	log := setupLogger(fn, r.Context(), h.log)
	log.Info("start creating user", slog.String("url", r.URL.String()))
	var user models.User
	err := render.DecodeJSON(r.Body, &user)
	if err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massege": "invalid JSON payload",
		})
		return
	}
	name := strings.TrimSpace(user.Name)
	email := strings.ToLower(strings.TrimSpace(user.Email))
	validationErrors := validateUser(name, email)
	if validationErrors != nil {
		log.Error("validation error")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, validationErrors)
		return
	}
	user.Name, user.Email = name, email
	id, err := h.storage.CreateUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, postgres.ErrInvalidInput) {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid input"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"massege": "failed to create user"},
		)
		return
	}
	log.Info("user created successfully")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"status": "user created successfully",
		"id":     id,
		"user":   user,
	})

}

func (h Hendler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.GetUserByEmail"
	log := setupLogger(fn, r.Context(), h.log)
	log.Info("start getting user", slog.String("url", r.URL.String()))

	email := chi.URLParam(r, "email")
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "bad request",
			"massage": "email is empty",
		})
		return
	}
	user, err := h.storage.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			log.Info("email is not found", slog.String("url", r.URL.String()))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"massage": "not found",
			})
			return
		}
		log.Error("error getting user by email", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "internal server error",
			"massage": "failed getting user",
		})
		return
	}
	log.Info("getting user by email is successfully",
		slog.String("url", r.URL.String()),
		slog.String("name", user.Name))
	render.JSON(w, r, user)

}

func (h Hendler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.GetAllUsers"
	log := setupLogger(fn, r.Context(), h.log)
	log.Info("start getting all user", slog.String("URL", r.URL.String()))
	users, err := h.storage.GetAllUsers(r.Context())
	if err != nil {
		slog.Error("error getting all users", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "interal server error",
			"massege": "error getting all users",
		})
		return
	}
	log.Info("getting all user is successfully",
		slog.String("url", r.URL.String()))
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]any{
		"status": "getting all users is successfully",
		"users":  users,
	})

}
