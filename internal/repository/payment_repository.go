package repository

import (
	"context"
	"time"

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

func (r *PaymentRepository) CountTablesAwaitingPayment(ctx context.Context, branchID string) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx, `
		SELECT COUNT(DISTINCT t.id)
		FROM payment_requests p
		JOIN sessions s ON s.id = p.session_id
		JOIN tables t ON t.id = s.table_id
		WHERE t.branch_id = $1 AND p.status != 'cancelled' AND s.status != 'closed'
	`, branchID).Scan(&count)
	return count, err
}

// ListByBranch returns payment requests for a branch. When from/to are nil, it
// returns only currently-pending requests (session not yet closed) — the
// "needs action now" view. When from/to are provided, it returns historical
// requests within that requested_at range regardless of session status.
func (r *PaymentRepository) ListByBranch(ctx context.Context, branchID string, from, to *time.Time) ([]models.PaymentRequest, error) {
	query := `
		SELECT p.id, p.session_id, p.status, p.requested_at, p.confirmed_at, t.code, t.area,
			(SELECT MIN(o.created_at) FROM orders o WHERE o.session_id = p.session_id) AS ordered_at,
			COALESCE((SELECT SUM(o.total) FROM orders o WHERE o.session_id = p.session_id AND o.status != 'cancelled'), 0)::text AS total
		FROM payment_requests p
		JOIN sessions s ON s.id = p.session_id
		JOIN tables t ON t.id = s.table_id
		WHERE t.branch_id = $1 AND p.status != 'cancelled'
	`
	args := []any{branchID}
	if from != nil && to != nil {
		query += ` AND p.requested_at >= $2 AND p.requested_at <= $3`
		args = append(args, *from, *to)
	} else {
		query += ` AND s.status != 'closed'`
	}
	query += ` ORDER BY p.requested_at ASC`

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.PaymentRequest
	for rows.Next() {
		var p models.PaymentRequest
		if err := rows.Scan(&p.ID, &p.SessionID, &p.Status, &p.RequestedAt, &p.ConfirmedAt, &p.TableCode, &p.TableArea, &p.OrderedAt, &p.Total); err != nil {
			return nil, err
		}
		requests = append(requests, p)
	}
	return requests, rows.Err()
}
