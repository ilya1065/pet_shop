package models

import "time"

type User struct {
	ID    int    `json:"ID"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Product struct {
	ID    int
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"` // количество на складе
}

type Customer struct {
	ID    int
	Name  string
	Email string
}

type Order struct {
	ID         int       `json:"id"`
	CustomerID int       `json:"customer_id"`
	CreatedAt  time.Time `json:"created_at"`
	TotalPrice float64   `json:"total_price"`
}

type OrderItem struct {
	ID        int `json:"id"`
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Transactions struct {
	ID       int
	OrderID  int
	Status   string
	Amount   float64
	CreateAt time.Time
}

type OrderDetail struct {
	OrderID            int         `json:"order_id"`
	CreateAt           time.Time   `json:"create_at"`
	TotalPrice         float64     `json:"total_price"`
	Items              []OrderItem `json:"items"`
	TransactionsStatus string      `json:"transactions_status"`
}

type PopularProduct struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	SumSold     float64 `json:"sum_sold"`
}
