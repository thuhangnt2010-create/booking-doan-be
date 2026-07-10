package models

import "time"

type StaffCallRequest struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	TableCode string    `json:"tableCode,omitempty"`
	TableArea string    `json:"tableArea,omitempty"`
}
