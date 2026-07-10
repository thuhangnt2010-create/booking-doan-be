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

func (r *TableRepository) Create(ctx context.Context, branchID, area, code string) (*models.Table, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO tables (branch_id, area, code, status)
		VALUES ($1, $2, $3, 'ready')
		RETURNING id, branch_id, area, code, status
	`, branchID, area, code)
	var t models.Table
	if err := row.Scan(&t.ID, &t.BranchID, &t.Area, &t.Code, &t.Status); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TableRepository) ListByBranch(ctx context.Context, branchID string) ([]models.TableWithQR, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT t.id, t.branch_id, t.area, t.code, t.status, q.token
		FROM tables t
		LEFT JOIN qr_codes q ON q.table_id = t.id AND q.active = true
		WHERE t.branch_id = $1
		ORDER BY t.area, t.code
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []models.TableWithQR
	for rows.Next() {
		var t models.TableWithQR
		var token *string
		if err := rows.Scan(&t.ID, &t.BranchID, &t.Area, &t.Code, &t.Status, &token); err != nil {
			return nil, err
		}
		if token != nil {
			t.ActiveQRToken = *token
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}
