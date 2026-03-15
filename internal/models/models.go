package models

import (
	"encoding/json"
	"os"
	"sync"
)

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
	Mu          sync.RWMutex
	Products    map[int]*Product
	Orders      map[int]*Order
	NextOrderID int
}

type StoreData struct {
	Products    map[int]*Product
	Orders      map[int]*Order
	NextOrderID int
}

func NewStore() *Store {
	return &Store{
		Products:    make(map[int]*Product),
		Orders:      make(map[int]*Order),
		NextOrderID: 0,
	}
}

func (store *Store) SaveStore(filename string) error {
	store.Mu.RLock()
	defer store.Mu.RUnlock()

	data := StoreData{
		Products:    store.Products,
		Orders:      store.Orders,
		NextOrderID: store.NextOrderID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func LoadStore(filename string) (*Store, error) {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data StoreData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return &Store{
		Products:    data.Products,
		Orders:      data.Orders,
		NextOrderID: data.NextOrderID,
	}, nil
}
