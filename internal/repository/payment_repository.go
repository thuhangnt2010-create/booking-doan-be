package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type PaymentRepository struct {
	DB *pgxpool.Pool
}

func (r *PaymentRepository) Create(ctx context.Context, sessionID string) (*models.PaymentRequest, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO payment_requests (session_id, status)
		VALUES ($1, 'requested')
		RETURNING id, session_id, status, requested_at, confirmed_at
	`, sessionID)

	var p models.PaymentRequest
	if err := row.Scan(&p.ID, &p.SessionID, &p.Status, &p.RequestedAt, &p.ConfirmedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) Confirm(ctx context.Context, id string) (string, error) {
	row := r.DB.QueryRow(ctx, `
		UPDATE payment_requests SET status = 'confirmed', confirmed_at = NOW()
		WHERE id = $1
		RETURNING session_id
	`, id)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return sessionID, nil
}
