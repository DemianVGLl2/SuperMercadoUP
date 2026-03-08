package models

import "sync"

type Product struct {
	ID    int
	Name  string
	Price float64
	Stock int
}

type OrderItem struct {
	ProductID int
	Quantity  int
	UnitPrice float64
}

type OrderStatus string

const (
	Created   OrderStatus = "CREATED"
	Completed OrderStatus = "COMPLETED"
	Cancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID     int
	Items  []OrderItem
	Total  float64
	Status OrderStatus
}

type Cart struct {
	Items []OrderItem
}

type Store struct {
	mu       sync.RWMutex
	Products map[int]*Product
	Orders   map[int]*Order
}

func NewStore() *Store {
	return &Store{
		Products: make(map[int]*Product),
		Orders:   make(map[int]*Order),
	}
}
