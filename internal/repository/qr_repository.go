package repository

import (
	"context"

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
