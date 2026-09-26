package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	const fn = "storage.postgres.orders.CreateOrder"
	var id int
	err := s.db.QueryRow(ctx, `insert into orders(user_id,total_price,Create_at)
									values ($1,0,current_timestamp)
									returning id`).Scan(&id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("%w, %s", err, fn)
	}
	return id, nil
}

func (s Storage) AddOrderItem(ctx context.Context, item models.OrderItem) error {
	const fn = "storage.postgres.orders.AddOrederItem"
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%w, %s", err, fn)
	}
	defer tx.Rollback(ctx)
	var price int
	err = tx.QueryRow(ctx, `select price from products where id = $1`, item.ProductID).Scan(&price)
	if err != nil {
		return fmt.Errorf("%w,%s", err, fn)
	}
	_, err = tx.Exec(ctx, `insert into order_items(order_id,product_id,quantity) values ($1,$2,$3)`, item.OrderID, item.ProductID, item.Quantity)
	if err != nil {
		return fmt.Errorf("%w, %s", err, fn)
	}
	_, err = tx.Exec(ctx, `update orders set total_price = total_price + $1 * $2  where id = $3`, price,item.Quantity, item.OrderID)
	if err != nil {
		return fmt.Errorf("%w, %s", err, fn)
	}
	 err = tx.Commit(ctx)
	 if err != nil {
		 return fmt.Errorf("%w, %s", err, fn)
	 }
	return nil
}

func (s Storage) GetOrderByID(ctx context.Context, id int) (*models.Order, error) {
	const fn = "storage.postgers.order.GetOrderByID"
	var order models.Order
	err := s.db.QueryRow(ctx, `select id, user_id,create_at from orders where id = $1`, id
	).Scan(&order.ID,&order.CustomerID,&order.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", err, fn)
	}

	return &order, nil

}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	const fn = "storage.postgers.order.GetOrderByUserEmail"
	rows,err := s.db.Query(ctx,`select id , user_id,cretaed_at 
									from orders as o 
									join users as u 
									on  u.id = o.user_id
									where u.email = $1`,email)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", err, fn)
	}
	var orders []models.Order
	for rows.Next(){
		var order models.Order
		err = rows.Scan(order.ID,order.CustomerID, order.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%w, %s", err, fn)
		}
		orders = append(orders, order)

	}
	return orders, nil
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgers.order.GetOrderItemsByOrderID"
	var items []models.OrderItem
	rows , err := s.db.Query(ctx,`select `)



	return nil, nil
}
