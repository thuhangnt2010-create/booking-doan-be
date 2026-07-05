package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type StaffCallRepository struct {
	DB *pgxpool.Pool
}

func (r *StaffCallRepository) Create(ctx context.Context, sessionID, callType string) (*models.StaffCallRequest, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO staff_call_requests (session_id, type, status)
		VALUES ($1, $2, 'sent')
		RETURNING id, session_id, type, status, created_at
	`, sessionID, callType)

	var c models.StaffCallRequest
	if err := row.Scan(&c.ID, &c.SessionID, &c.Type, &c.Status, &c.CreatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *StaffCallRepository) UpdateStatus(ctx context.Context, id, status string) (string, error) {
	row := r.DB.QueryRow(ctx, `UPDATE staff_call_requests SET status = $1 WHERE id = $2 RETURNING session_id`, status, id)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return sessionID, nil
}

func (r *StaffCallRepository) ListBySession(ctx context.Context, sessionID string) ([]models.StaffCallRequest, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, session_id, type, status, created_at
		FROM staff_call_requests
		WHERE session_id = $1
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []models.StaffCallRequest
	for rows.Next() {
		var c models.StaffCallRequest
		if err := rows.Scan(&c.ID, &c.SessionID, &c.Type, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, rows.Err()
}
