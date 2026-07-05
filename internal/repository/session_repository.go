package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type SessionRepository struct {
	DB *pgxpool.Pool
}

func (r *SessionRepository) FindActiveByTable(ctx context.Context, tableID string) (*models.Session, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, table_id, status, started_at, ended_at
		FROM sessions
		WHERE table_id = $1 AND status IN ('active', 'payment_requested')
		ORDER BY started_at DESC
		LIMIT 1
	`, tableID)

	var s models.Session
	err := row.Scan(&s.ID, &s.TableID, &s.Status, &s.StartedAt, &s.EndedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) Create(ctx context.Context, tableID string) (*models.Session, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO sessions (table_id, status)
		VALUES ($1, 'active')
		RETURNING id, table_id, status, started_at, ended_at
	`, tableID)

	var s models.Session
	if err := row.Scan(&s.ID, &s.TableID, &s.Status, &s.StartedAt, &s.EndedAt); err != nil {
		return nil, err
	}
	return &s, nil
}
