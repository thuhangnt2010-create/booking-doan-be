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
	return r.UpdateItemFull(ctx, id, UpdateItemInput{Price: price, Status: status})
}

type UpdateItemInput struct {
	CategoryID      *string
	Code            *string
	Name            *string
	Price           *string
	Unit            *string
	PrepTimeMinutes *int
	Description     *string
	Ingredients     *string
	AllergyInfo     *string
	IsPromo         *bool
	IsBestSeller    *bool
	IsNew           *bool
	Status          *string
}

// UpdateItemFull updates any subset of an item's fields and returns the item's branch id for realtime broadcast.
func (r *MenuRepository) UpdateItemFull(ctx context.Context, id string, in UpdateItemInput) (string, error) {
	var setClauses []string
	var args []any

	add := func(col string, val any) {
		args = append(args, val)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if in.CategoryID != nil {
		add("category_id", *in.CategoryID)
	}
	if in.Code != nil {
		add("code", *in.Code)
	}
	if in.Name != nil {
		add("name", *in.Name)
	}
	if in.Price != nil {
		add("price", *in.Price)
	}
	if in.Unit != nil {
		add("unit", *in.Unit)
	}
	if in.PrepTimeMinutes != nil {
		add("prep_time_minutes", *in.PrepTimeMinutes)
	}
	if in.Description != nil {
		add("description", *in.Description)
	}
	if in.Ingredients != nil {
		add("ingredients", *in.Ingredients)
	}
	if in.AllergyInfo != nil {
		add("allergy_info", *in.AllergyInfo)
	}
	if in.IsPromo != nil {
		add("is_promo", *in.IsPromo)
	}
	if in.IsBestSeller != nil {
		add("is_best_seller", *in.IsBestSeller)
	}
	if in.IsNew != nil {
		add("is_new", *in.IsNew)
	}
	if in.Status != nil {
		add("status", *in.Status)
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

	return r.itemBranchID(ctx, id)
}

func (r *MenuRepository) itemBranchID(ctx context.Context, id string) (string, error) {
	var branchID string
	row := r.DB.QueryRow(ctx, `
		SELECT c.branch_id FROM menu_items i JOIN menu_categories c ON c.id = i.category_id WHERE i.id = $1
	`, id)
	if err := row.Scan(&branchID); err != nil {
		return "", err
	}
	return branchID, nil
}

type CreateItemInput struct {
	CategoryID      string
	Code            string
	Name            string
	Price           string
	Unit            string
	PrepTimeMinutes int
	Description     string
	Ingredients     string
	AllergyInfo     string
	IsPromo         bool
	IsBestSeller    bool
	IsNew           bool
	Status          string
}

// CreateItem inserts a new menu item and returns its id and branch id for realtime broadcast.
func (r *MenuRepository) CreateItem(ctx context.Context, in CreateItemInput) (id string, branchID string, err error) {
	status := in.Status
	if status == "" {
		status = "available"
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO menu_items (category_id, code, name, price, unit, prep_time_minutes, description, ingredients, allergy_info, is_promo, is_best_seller, is_new, status, image_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, '')
		RETURNING id
	`, in.CategoryID, in.Code, in.Name, in.Price, in.Unit, in.PrepTimeMinutes, in.Description, in.Ingredients, in.AllergyInfo, in.IsPromo, in.IsBestSeller, in.IsNew, status)
	if err := row.Scan(&id); err != nil {
		return "", "", err
	}
	branchID, err = r.itemBranchID(ctx, id)
	return id, branchID, err
}

// DeleteItem removes a menu item and returns its branch id for realtime broadcast.
func (r *MenuRepository) DeleteItem(ctx context.Context, id string) (string, error) {
	branchID, err := r.itemBranchID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	if _, err := r.DB.Exec(ctx, `DELETE FROM menu_items WHERE id = $1`, id); err != nil {
		return "", err
	}
	return branchID, nil
}

func (r *MenuRepository) ListCategories(ctx context.Context, branchID string) ([]models.MenuCategory, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, branch_id, name, position FROM menu_categories WHERE branch_id = $1 ORDER BY position, name
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.MenuCategory
	for rows.Next() {
		var c models.MenuCategory
		if err := rows.Scan(&c.ID, &c.BranchID, &c.Name, &c.Position); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *MenuRepository) CreateCategory(ctx context.Context, branchID, name string, position int) (*models.MenuCategory, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO menu_categories (branch_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, branch_id, name, position
	`, branchID, name, position)
	var c models.MenuCategory
	if err := row.Scan(&c.ID, &c.BranchID, &c.Name, &c.Position); err != nil {
		return nil, err
	}
	return &c, nil
}
