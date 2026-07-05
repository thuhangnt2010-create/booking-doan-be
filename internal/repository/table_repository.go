package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TableRepository struct {
	DB *pgxpool.Pool
}

func (r *TableRepository) SetStatus(ctx context.Context, tableID, status string) error {
	_, err := r.DB.Exec(ctx, `UPDATE tables SET status = $1 WHERE id = $2`, status, tableID)
	return err
}
