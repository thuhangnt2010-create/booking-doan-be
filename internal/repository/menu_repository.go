package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type MenuRepository struct {
	DB *pgxpool.Pool
}

type MenuFilter struct {
	BranchID   string
	Search     string
	CategoryID string
	MinPrice   string
	MaxPrice   string
	Promo      bool
	BestSeller bool
	IsNew      bool
	Sort       string
}

func (r *MenuRepository) ListItems(ctx context.Context, f MenuFilter) ([]models.MenuItem, error) {
	query := `
		SELECT i.id, i.category_id, c.name, i.code, i.name, i.price::text, i.status,
		       i.unit, i.prep_time_minutes, i.is_promo, i.is_best_seller, i.is_new,
		       i.image_key, i.description, i.created_at
		FROM menu_items i
		JOIN menu_categories c ON c.id = i.category_id
		WHERE c.branch_id = $1`
	args := []any{f.BranchID}

	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		query += fmt.Sprintf(" AND (i.name ILIKE $%d OR i.code ILIKE $%d)", len(args), len(args))
	}
	if f.CategoryID != "" {
		args = append(args, f.CategoryID)
		query += fmt.Sprintf(" AND i.category_id = $%d", len(args))
	}
	if f.MinPrice != "" {
		args = append(args, f.MinPrice)
		query += fmt.Sprintf(" AND i.price >= $%d", len(args))
	}
	if f.MaxPrice != "" {
		args = append(args, f.MaxPrice)
		query += fmt.Sprintf(" AND i.price <= $%d", len(args))
	}
	if f.Promo {
		query += " AND i.is_promo = true"
	}
	if f.BestSeller {
		query += " AND i.is_best_seller = true"
	}
	if f.IsNew {
		query += " AND i.is_new = true"
	}

	switch f.Sort {
	case "price_asc":
		query += " ORDER BY i.price ASC"
	case "price_desc":
		query += " ORDER BY i.price DESC"
	case "best_seller":
		query += " ORDER BY i.is_best_seller DESC, i.created_at DESC"
	case "newest":
		query += " ORDER BY i.created_at DESC"
	default:
		query += " ORDER BY c.position ASC, i.created_at ASC"
	}

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.MenuItem
	for rows.Next() {
		var m models.MenuItem
		if err := rows.Scan(
			&m.ID, &m.CategoryID, &m.CategoryName, &m.Code, &m.Name, &m.Price, &m.Status,
			&m.Unit, &m.PrepTimeMinutes, &m.IsPromo, &m.IsBestSeller, &m.IsNew,
			&m.ImageKey, &m.Description, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *MenuRepository) GetItemDetail(ctx context.Context, itemID string) (*models.MenuItemDetail, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT i.id, i.category_id, c.name, i.code, i.name, i.price::text, i.status,
		       i.unit, i.prep_time_minutes, i.is_promo, i.is_best_seller, i.is_new,
		       i.image_key, i.description, i.created_at, i.ingredients, i.allergy_info
		FROM menu_items i
		JOIN menu_categories c ON c.id = i.category_id
		WHERE i.id = $1
	`, itemID)

	var d models.MenuItemDetail
	err := row.Scan(
		&d.ID, &d.CategoryID, &d.CategoryName, &d.Code, &d.Name, &d.Price, &d.Status,
		&d.Unit, &d.PrepTimeMinutes, &d.IsPromo, &d.IsBestSeller, &d.IsNew,
		&d.ImageKey, &d.Description, &d.CreatedAt, &d.Ingredients, &d.AllergyInfo,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	rows, err := r.DB.Query(ctx, `
		SELECT id, item_id, type, name, price_delta::text
		FROM menu_item_options
		WHERE item_id = $1
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.MenuItemOption
		if err := rows.Scan(&o.ID, &o.ItemID, &o.Type, &o.Name, &o.PriceDelta); err != nil {
			return nil, err
		}
		d.Options = append(d.Options, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &d, nil
}

// UpdateItem updates price and/or status and returns the item's branch id for realtime broadcast.
func (r *MenuRepository) UpdateItem(ctx context.Context, id string, price, status *string) (string, error) {
	var setClauses []string
	var args []any

	if price != nil {
		args = append(args, *price)
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", len(args)))
	}
	if status != nil {
		args = append(args, *status)
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if len(setClauses) == 0 {
		return "", errors.New("no fields to update")
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE menu_items SET %s WHERE id = $%d", strings.Join(setClauses, ", "), len(args))
	if ct, err := r.DB.Exec(ctx, query, args...); err != nil {
		return "", err
	} else if ct.RowsAffected() == 0 {
		return "", ErrNotFound
	}

	var branchID string
	row := r.DB.QueryRow(ctx, `
		SELECT c.branch_id FROM menu_items i JOIN menu_categories c ON c.id = i.category_id WHERE i.id = $1
	`, id)
	if err := row.Scan(&branchID); err != nil {
		return "", err
	}
	return branchID, nil
}
