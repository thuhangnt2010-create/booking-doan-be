package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type QRRepository struct {
	DB *pgxpool.Pool
}

type ResolvedQR struct {
	Table      models.Table
	Branch     models.Branch
	Restaurant models.Restaurant
}

func (r *QRRepository) FindActiveByToken(ctx context.Context, token string) (*ResolvedQR, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT t.id, t.branch_id, t.area, t.code, t.status,
		       b.id, b.restaurant_id, b.name, b.status,
		       rs.id, rs.name
		FROM qr_codes q
		JOIN tables t ON t.id = q.table_id
		JOIN branches b ON b.id = t.branch_id
		JOIN restaurants rs ON rs.id = b.restaurant_id
		WHERE q.token = $1 AND q.active = true
	`, token)

	var res ResolvedQR
	err := row.Scan(
		&res.Table.ID, &res.Table.BranchID, &res.Table.Area, &res.Table.Code, &res.Table.Status,
		&res.Branch.ID, &res.Branch.RestaurantID, &res.Branch.Name, &res.Branch.Status,
		&res.Restaurant.ID, &res.Restaurant.Name,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *QRRepository) CreateForTable(ctx context.Context, tableID string) (*models.QRCode, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE qr_codes SET active = false WHERE table_id = $1 AND active = true`, tableID); err != nil {
		return nil, err
	}

	token := generateQRToken()
	row := tx.QueryRow(ctx, `
		INSERT INTO qr_codes (table_id, token, active)
		VALUES ($1, $2, true)
		RETURNING id, table_id, token, active, created_at
	`, tableID, token)

	var q models.QRCode
	if err := row.Scan(&q.ID, &q.TableID, &q.Token, &q.Active, &q.CreatedAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QRRepository) GetActiveByTable(ctx context.Context, tableID string) (*models.QRCode, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, table_id, token, active, created_at
		FROM qr_codes WHERE table_id = $1 AND active = true
	`, tableID)
	var q models.QRCode
	if err := row.Scan(&q.ID, &q.TableID, &q.Token, &q.Active, &q.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &q, nil
}

func generateQRToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
