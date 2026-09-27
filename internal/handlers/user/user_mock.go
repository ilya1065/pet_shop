package user

import (
	"context"
	"go-pet-shop/internal/models"
)

type Users_mock interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}
