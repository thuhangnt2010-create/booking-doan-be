package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type TableRepository struct {
	DB *pgxpool.Pool
}

func (r *TableRepository) SetStatus(ctx context.Context, tableID, status string) error {
	_, err := r.DB.Exec(ctx, `UPDATE tables SET status = $1 WHERE id = $2`, status, tableID)
	return err
}

func (r *TableRepository) FindByID(ctx context.Context, tableID string) (*models.Table, error) {
	row := r.DB.QueryRow(ctx, `SELECT id, branch_id, area, code, status FROM tables WHERE id = $1`, tableID)
	var t models.Table
	if err := row.Scan(&t.ID, &t.BranchID, &t.Area, &t.Code, &t.Status); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}
