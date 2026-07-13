package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

const (
	vatRate       = 0.10
	maxNoteLength = 200
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionClosed   = errors.New("session is not active")
	ErrEmptyOrder      = errors.New("order must have at least one item")
)

type OrderValidationError struct {
	Code    string
	Message string
}

func (e *OrderValidationError) Error() string { return e.Message }

type OrderService struct {
	Session *repository.SessionRepository
	Table   *repository.TableRepository
	Menu    *repository.MenuRepository
	Order   *repository.OrderRepository
	Hub     *realtime.Hub
}

type OrderItemRequest struct {
	ItemID    string
	Qty       int
	Note      string
	OptionIDs []string
}

type CreateOrderRequest struct {
	SessionID       string
	ClientRequestID string
	Items           []OrderItemRequest
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*models.Order, error) {
	if req.ClientRequestID != "" {
		existing, err := s.Order.FindByClientRequestID(ctx, req.ClientRequestID)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}

	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	session, err := s.Session.FindByID(ctx, req.SessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if session.Status != "active" {
		return nil, ErrSessionClosed
	}

	var resolved []repository.OrderItemInput
	var subtotal float64

	for _, itReq := range req.Items {
		if itReq.Qty <= 0 {
			return nil, &OrderValidationError{Code: "INVALID_QTY", Message: "Số lượng phải lớn hơn 0"}
		}
		if len(itReq.Note) > maxNoteLength {
			return nil, &OrderValidationError{Code: "NOTE_TOO_LONG", Message: fmt.Sprintf("Ghi chú tối đa %d ký tự", maxNoteLength)}
		}

		detail, err := s.Menu.GetItemDetail(ctx, itReq.ItemID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, &OrderValidationError{Code: "ITEM_NOT_FOUND", Message: fmt.Sprintf("Món %s không tồn tại", itReq.ItemID)}
			}
			return nil, err
		}
		if detail.Status == "out_of_stock" || detail.Status == "suspended" {
			return nil, &OrderValidationError{Code: "ITEM_UNAVAILABLE", Message: fmt.Sprintf("Món %s hiện không thể đặt (%s)", detail.Name, detail.Status)}
		}

		unitPrice, err := strconv.ParseFloat(detail.Price, 64)
		if err != nil {
			return nil, err
		}

		var opts []repository.OrderItemOptionInput
		for _, optID := range itReq.OptionIDs {
			found := false
			for _, o := range detail.Options {
				if o.ID == optID {
					delta, _ := strconv.ParseFloat(o.PriceDelta, 64)
					unitPrice += delta
					opts = append(opts, repository.OrderItemOptionInput{OptionID: o.ID, Name: o.Name, PriceDelta: delta})
					found = true
					break
				}
			}
			if !found {
				return nil, &OrderValidationError{Code: "OPTION_INVALID", Message: fmt.Sprintf("Option %s không hợp lệ cho món %s", optID, detail.Name)}
			}
		}

		subtotal += unitPrice * float64(itReq.Qty)

		resolved = append(resolved, repository.OrderItemInput{
			ItemID: itReq.ItemID, ItemName: detail.Name, Qty: itReq.Qty,
			UnitPrice: unitPrice, Note: itReq.Note, Options: opts,
		})
	}

	vat := subtotal * vatRate
	total := subtotal + vat
	code := generateOrderCode()

	order, err := s.Order.Create(ctx, req.SessionID, code, req.ClientRequestID, subtotal, vat, total, resolved)
	if err != nil {
		return nil, err
	}

	s.broadcastNewOrder(ctx, session, order)

	return order, nil
}

var validStatuses = map[string]bool{
	"sent": true, "received": true, "cooking": true, "done": true, "served": true, "cancelled": true,
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	if !validStatuses[status] {
		return &OrderValidationError{Code: "INVALID_STATUS", Message: "Trạng thái không hợp lệ"}
	}
	if status == "cancelled" {
		order, err := s.Order.GetByID(ctx, orderID)
		if err != nil {
			return err
		}
		for _, it := range order.Items {
			if it.Status == "served" {
				return &OrderValidationError{Code: "ORDER_HAS_SERVED_ITEMS", Message: "Không thể hủy order đã có món giao cho khách"}
			}
		}
	}
	sessionID, err := s.Order.UpdateStatus(ctx, orderID, status)
	if err != nil {
		return err
	}

	payload := fmt.Sprintf(`{"type":"order_status","orderId":"%s","status":"%s"}`, orderID, status)
	s.broadcastToSessionAndBranch(ctx, sessionID, payload)
	return nil
}

func (s *OrderService) UpdateOrderItemStatus(ctx context.Context, orderItemID, status string) error {
	if !validStatuses[status] {
		return &OrderValidationError{Code: "INVALID_STATUS", Message: "Trạng thái không hợp lệ"}
	}
	orderID, err := s.Order.UpdateItemStatus(ctx, orderItemID, status)
	if err != nil {
		return err
	}

	order, err := s.Order.GetByID(ctx, orderID)
	if err != nil {
		return nil
	}

	payload := fmt.Sprintf(`{"type":"order_item_status","orderItemId":"%s","status":"%s"}`, orderItemID, status)
	s.broadcastToSessionAndBranch(ctx, order.SessionID, payload)
	return nil
}

func (s *OrderService) broadcastToSessionAndBranch(ctx context.Context, sessionID, payload string) {
	if s.Hub == nil {
		return
	}
	session, err := s.Session.FindByID(ctx, sessionID)
	if err != nil {
		return
	}
	table, err := s.Table.FindByID(ctx, session.TableID)
	if err != nil {
		return
	}
	s.Hub.Broadcast("orders:branch:"+table.BranchID, []byte(payload))
	s.Hub.Broadcast("orders:session:"+sessionID, []byte(payload))
}

func (s *OrderService) broadcastNewOrder(ctx context.Context, session *models.Session, order *models.Order) {
	payload := fmt.Sprintf(`{"type":"order_created","orderId":"%s","code":"%s"}`, order.ID, order.Code)
	s.broadcastToSessionAndBranch(ctx, session.ID, payload)
}

func generateOrderCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("ORD%s%s", time.Now().Format("060102"), hex.EncodeToString(b))
}
