package models

import "time"


type User struct {
	ID int `json:"ID"`
	Name string `json:"name"`
	Email string `json:"email"`
}

type Product struct {
	ID    int
	Name  string `json:"name"`
	Price float64 `json:"price"`
	Stock int `json:"stock"` // количество на складе
}

type Customer struct {
	ID    int
	Name  string
	Email string
}

type Order struct {
	ID         int
	CustomerID int
	CreatedAt  time.Time
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
}
