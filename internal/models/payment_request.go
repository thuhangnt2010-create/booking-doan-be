package models

import "time"

type PaymentRequest struct {
	ID          string     `json:"id"`
	SessionID   string     `json:"sessionId"`
	Status      string     `json:"status"`
	RequestedAt time.Time  `json:"requestedAt"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`
	TableCode   string     `json:"tableCode,omitempty"`
	TableArea   string     `json:"tableArea,omitempty"`
	OrderedAt   *time.Time `json:"orderedAt,omitempty"`
	Total       string     `json:"total"`
}

type Bill struct {
	SessionID string  `json:"sessionId"`
	Orders    []Order `json:"orders"`
	Subtotal  string  `json:"subtotal"`
	VATAmount string  `json:"vatAmount"`
	Total     string  `json:"total"`
}
