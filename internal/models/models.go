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
