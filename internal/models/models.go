package models

import "time"

type Restaurant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Branch struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurantId"`
	Name         string `json:"name"`
	Status       string `json:"status"`
}

type Table struct {
	ID       string `json:"id"`
	BranchID string `json:"branchId"`
	Area     string `json:"area"`
	Code     string `json:"code"`
	Status   string `json:"status"`
}

type Session struct {
	ID        string     `json:"id"`
	TableID   string     `json:"tableId"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

type MenuCategory struct {
	ID       string `json:"id"`
	BranchID string `json:"branchId"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type MenuItem struct {
	ID              string    `json:"id"`
	CategoryID      string    `json:"categoryId"`
	CategoryName    string    `json:"categoryName"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Price           string    `json:"price"`
	Status          string    `json:"status"`
	Unit            string    `json:"unit"`
	PrepTimeMinutes int       `json:"prepTimeMinutes"`
	IsPromo         bool      `json:"isPromo"`
	IsBestSeller    bool      `json:"isBestSeller"`
	IsNew           bool      `json:"isNew"`
	ImageKey        string    `json:"imageKey"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"createdAt"`
}

type MenuItemDetail struct {
	MenuItem
	Ingredients  string           `json:"ingredients"`
	AllergyInfo  string           `json:"allergyInfo"`
	Options      []MenuItemOption `json:"options"`
}

type MenuItemOption struct {
	ID         string `json:"id"`
	ItemID     string `json:"itemId"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	PriceDelta string `json:"priceDelta"`
}

type Order struct {
	ID              string      `json:"id"`
	SessionID       string      `json:"sessionId"`
	Code            string      `json:"code"`
	Status          string      `json:"status"`
	Subtotal        string      `json:"subtotal"`
	VATAmount       string      `json:"vatAmount"`
	Total           string      `json:"total"`
	CreatedAt       time.Time   `json:"createdAt"`
	Items           []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        string              `json:"id"`
	OrderID   string              `json:"orderId"`
	ItemID    string              `json:"itemId"`
	ItemName  string              `json:"itemName"`
	Qty       int                 `json:"qty"`
	UnitPrice string              `json:"unitPrice"`
	Note      string              `json:"note"`
	Status    string              `json:"status"`
	Options   []OrderItemOption   `json:"options"`
}

type OrderItemOption struct {
	ID          string `json:"id"`
	OrderItemID string `json:"orderItemId"`
	Name        string `json:"name"`
	PriceDelta  string `json:"priceDelta"`
}
