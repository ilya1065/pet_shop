package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	const fn = "storage.postgres.orders.CreateOrder"
	var id int
	err := s.db.QueryRow(ctx, `insert into orders(user_id,total_price,created_at)
									values ($1,0,current_timestamp)
									returning id`, order.CustomerID).Scan(&id)
	if err != nil {
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
	_, err = tx.Exec(ctx, `update orders set total_price = total_price +($1::integer * $2::integer)  where id = $3`, price, item.Quantity, item.OrderID)
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
	err := s.db.QueryRow(ctx, `select id, user_id, created_at, total_price from orders where id = $1`, id).Scan(&order.ID, &order.CustomerID, &order.CreatedAt, &order.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", err, fn)
	}

	return &order, nil

}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.OrderWithItems, error) {
	const fn = "storage.postgers.order.GetOrderByUserEmail"
	rows, err := s.db.Query(ctx, `select o.id,
										o.user_id,
										o.created_at,
										o.total_price,
										oi.id,
										oi.order_id,
										oi.product_id,
										oi.quantity
								 from orders as o
								 join users as u on u.id = o.user_id
								 left join order_items as oi on oi.order_id = o.id
								 where u.email = $1
								 order by o.id, oi.id`, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var orders []models.OrderWithItems
	for rows.Next() {
		var (
			order                                            models.Order
			item                                             models.OrderItem
			itemID, itemOrderID, itemProductID, itemQuantity pgtype.Int4
		)
		err = rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.CreatedAt,
			&order.TotalPrice,
			&itemID,
			&itemOrderID,
			&itemProductID,
			&itemQuantity,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		if len(orders) == 0 || orders[len(orders)-1].Order.ID != order.ID {
			orders = append(orders, models.OrderWithItems{
				Order: order,
				Items: make([]models.OrderItem, 0),
			})
		}

		if itemID.Valid {
			item.ID = int(itemID.Int32)
			item.OrderID = int(itemOrderID.Int32)
			item.ProductID = int(itemProductID.Int32)
			item.Quantity = int(itemQuantity.Int32)
			orders[len(orders)-1].Items = append(orders[len(orders)-1].Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return orders, nil
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgers.order.GetOrderItemsByOrderID"
	rows, err := s.db.Query(ctx, `select i.id , i.order_id, i.product_id,i.quantity from order_items as i 
										where i.order_id = $1`, orderID)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("%w, %s", err, fn)
	}

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err = rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("%w, %s", err, fn)
		}
		items = append(items, item)

	}

	return items, nil
}
