package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

type OrderItemOptionInput struct {
	OptionID   string
	Name       string
	PriceDelta float64
}

type OrderItemInput struct {
	ItemID    string
	ItemName  string
	Qty       int
	UnitPrice float64
	Note      string
	Options   []OrderItemOptionInput
}

func (r *OrderRepository) Create(ctx context.Context, sessionID, code, clientRequestID string, subtotal, vat, total float64, items []OrderItemInput) (*models.Order, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT id FROM sessions WHERE id = $1 FOR UPDATE`, sessionID); err != nil {
		return nil, err
	}
	var sequenceNo int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM orders WHERE session_id = $1`, sessionID).Scan(&sequenceNo); err != nil {
		return nil, err
	}

	var orderID string
	var reqID any = clientRequestID
	if clientRequestID == "" {
		reqID = nil
	}
	row := tx.QueryRow(ctx, `
		INSERT INTO orders (session_id, code, sequence_no, status, subtotal, vat_amount, total, client_request_id)
		VALUES ($1, $2, $3, 'sent', $4, $5, $6, $7)
		RETURNING id
	`, sessionID, code, sequenceNo, subtotal, vat, total, reqID)
	if err := row.Scan(&orderID); err != nil {
		return nil, err
	}

	for _, it := range items {
		var orderItemID string
		row := tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, item_id, qty, note, status, unit_price)
			VALUES ($1, $2, $3, $4, 'sent', $5)
			RETURNING id
		`, orderID, it.ItemID, it.Qty, it.Note, it.UnitPrice)
		if err := row.Scan(&orderItemID); err != nil {
			return nil, err
		}

		for _, opt := range it.Options {
			if _, err := tx.Exec(ctx, `
				INSERT INTO order_item_options (order_item_id, option_id, name, price_delta)
				VALUES ($1, $2, $3, $4)
			`, orderItemID, opt.OptionID, opt.Name, opt.PriceDelta); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.getFullOrder(ctx, orderID)
}

func (r *OrderRepository) getFullOrder(ctx context.Context, orderID string) (*models.Order, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, session_id, code, sequence_no, status, subtotal::text, vat_amount::text, total::text, created_at
		FROM orders WHERE id = $1
	`, orderID)

	var o models.Order
	if err := row.Scan(&o.ID, &o.SessionID, &o.Code, &o.SequenceNo, &o.Status, &o.Subtotal, &o.VATAmount, &o.Total, &o.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	rows, err := r.DB.Query(ctx, `
		SELECT oi.id, oi.order_id, oi.item_id, mi.name, oi.qty, oi.unit_price::text, oi.note, oi.status
		FROM order_items oi
		JOIN menu_items mi ON mi.id = oi.item_id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oi models.OrderItem
		if err := rows.Scan(&oi.ID, &oi.OrderID, &oi.ItemID, &oi.ItemName, &oi.Qty, &oi.UnitPrice, &oi.Note, &oi.Status); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, oi)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range o.Items {
		optRows, err := r.DB.Query(ctx, `
			SELECT id, order_item_id, name, price_delta::text
			FROM order_item_options
			WHERE order_item_id = $1
		`, o.Items[i].ID)
		if err != nil {
			return nil, err
		}
		for optRows.Next() {
			var opt models.OrderItemOption
			if err := optRows.Scan(&opt.ID, &opt.OrderItemID, &opt.Name, &opt.PriceDelta); err != nil {
				optRows.Close()
				return nil, err
			}
			o.Items[i].Options = append(o.Items[i].Options, opt)
		}
		optRows.Close()
	}

	return &o, nil
}

func (r *OrderRepository) FindByClientRequestID(ctx context.Context, clientRequestID string) (*models.Order, error) {
	row := r.DB.QueryRow(ctx, `SELECT id FROM orders WHERE client_request_id = $1`, clientRequestID)
	var orderID string
	if err := row.Scan(&orderID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.getFullOrder(ctx, orderID)
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID string) (*models.Order, error) {
	return r.getFullOrder(ctx, orderID)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID, status string) (string, error) {
	row := r.DB.QueryRow(ctx, `UPDATE orders SET status = $1 WHERE id = $2 RETURNING session_id`, status, orderID)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return sessionID, nil
}

func (r *OrderRepository) UpdateItemStatus(ctx context.Context, orderItemID, status string) (string, error) {
	row := r.DB.QueryRow(ctx, `UPDATE order_items SET status = $1 WHERE id = $2 RETURNING order_id`, status, orderItemID)
	var orderID string
	if err := row.Scan(&orderID); err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return orderID, nil
}

func (r *OrderRepository) CountTablesOrdering(ctx context.Context, branchID string) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx, `
		SELECT COUNT(DISTINCT t.id)
		FROM orders o
		JOIN sessions s ON s.id = o.session_id
		JOIN tables t ON t.id = s.table_id
		WHERE t.branch_id = $1 AND o.status NOT IN ('served', 'cancelled') AND s.status != 'closed'
	`, branchID).Scan(&count)
	return count, err
}

func (r *OrderRepository) ListByBranch(ctx context.Context, branchID string) ([]models.Order, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT o.id, t.code, t.area
		FROM orders o
		JOIN sessions s ON s.id = o.session_id
		JOIN tables t ON t.id = s.table_id
		WHERE t.branch_id = $1 AND o.status NOT IN ('served', 'cancelled') AND s.status != 'closed'
		ORDER BY o.created_at ASC
	`, branchID)
	if err != nil {
		return nil, err
	}
	type row struct{ id, code, area string }
	var list []row
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.id, &rr.code, &rr.area); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, rr)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var orders []models.Order
	for _, rr := range list {
		o, err := r.getFullOrder(ctx, rr.id)
		if err != nil {
			return nil, fmt.Errorf("load order %s: %w", rr.id, err)
		}
		o.TableCode = rr.code
		o.TableArea = rr.area
		orders = append(orders, *o)
	}
	return orders, nil
}

func (r *OrderRepository) ListBySession(ctx context.Context, sessionID string) ([]models.Order, error) {
	rows, err := r.DB.Query(ctx, `SELECT id FROM orders WHERE session_id = $1 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var orders []models.Order
	for _, id := range ids {
		o, err := r.getFullOrder(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("load order %s: %w", id, err)
		}
		orders = append(orders, *o)
	}
	return orders, nil
}
