package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateUser(ctx context.Context, user models.User) (int, error) {
	const fn = "storage.postgres.user.CreateUser"
	var id int
	err := s.db.QueryRow(ctx, `insert into users(name, email) values ($1, $2) returning id`, user.Name, user.Email).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", fn, ErrInvalidInput)
		}
		return 0, fmt.Errorf("%s,%w", fn, err)
	}
	return id, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {

	const fn = "storage.postgres.user.GetUserByEmail"
	var user models.User
	err := s.db.QueryRow(ctx, `select id, name, email from users where email = $1`, email).
		Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%s, %w", fn, err)
	}
	return &user, nil
}

func (s Storage) GetAllUsers(ctx context.Context) ([]models.User, error) {
	const fn = "storage.postgres.user.GetAllUsers"
	rows, err := s.db.Query(ctx, `select id, name ,email from users`)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", fn, err)
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", fn, err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", fn, err)
	}
	return users, nil
}
