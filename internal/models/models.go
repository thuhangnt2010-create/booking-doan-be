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
