package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
)

func (s *Storage) GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error) {
	const fn = "storage.postgers.analytics.GetUserOrderHistory"

	rows, err := s.db.Query(ctx, `
select o.id,
       o.created_at,
       o.total_price,
       oi.id,
       oi.product_id,
       oi.quantity,
       t.status
       from users as u 
join public.orders o on u.id = o.user_id
join public.order_items oi on o.id = oi.order_id
join public.transactions t on o.id = t.order_id
where u.email = $1`, email)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", err, fn)
	}
	ordersMap := make(map[int]*models.OrderDetail)
	for rows.Next() {
		var (
			order models.OrderDetail
			item  models.OrderItem
		)
		err = rows.Scan(
			&order.OrderID,
			&order.CreateAt,
			&order.TotalPrice,
			&item.ID,
			&item.ProductID,
			&item.Quantity,
			&order.TransactionsStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("%w, %s", err, fn)
		}

		existOrder, exist := ordersMap[order.OrderID]
		if !exist {
			existOrder = &order
			ordersMap[order.OrderID] = &order
		}

		existOrder.Items = append(existOrder.Items, item)
	}
	var ordersDetail []models.OrderDetail
	for _, orderDetail := range ordersMap {
		ordersDetail = append(ordersDetail, *orderDetail)
	}
	return ordersDetail, nil

}

func (s *Storage) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	const fn = "storage.postgers.analytics.GetPopularProduct"
	var popularProducts []models.PopularProduct
	rows, err := s.db.Query(ctx, `select p.id,
       							p.name,
       							sum(quantity* price) as SumSold,
       							sum(quantity) as TotalSold
						from order_items oi
						join public.products p on oi.product_id = p.id
    					join public.orders o on o.id = oi.order_id
						join public.transactions t on t.order_id = o.id
						where t.status = 'done'
						group by p.id
						order by TotalSold`)
	if err != nil {
		return nil, fmt.Errorf("%w,%s", err, fn)
	}
	defer rows.Close()
	for rows.Next() {
		var popularProduct models.PopularProduct
		err = rows.Scan(&popularProduct.ProductID,
			&popularProduct.ProductName,
			&popularProduct.SumSold,
			&popularProduct.TotalSold)
		if err != nil {
			return nil, fmt.Errorf("%w, %s", err, fn)
		}
		popularProducts = append(popularProducts, popularProduct)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w, %s", err, fn)
	}
	return popularProducts, nil
}
