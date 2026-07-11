package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type AdminUserRepository struct {
	DB *pgxpool.Pool
}

type AdminUserWithHash struct {
	models.AdminUser
	PasswordHash string
}

func (r *AdminUserRepository) FindByEmail(ctx context.Context, email string) (*AdminUserWithHash, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, email, password_hash, role, created_at
		FROM admin_users WHERE email = $1
	`, email)

	var u AdminUserWithHash
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
