package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

var ErrPaymentNotConfirmed = errors.New("payment request is not confirmed yet")

type PaymentService struct {
	Session *repository.SessionRepository
	Table   *repository.TableRepository
	Order   *repository.OrderRepository
	Payment *repository.PaymentRepository
	Hub     *realtime.Hub
}

func (s *PaymentService) RequestPayment(ctx context.Context, sessionID string) (*models.PaymentRequest, error) {
	session, err := s.Session.FindByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if session.Status != "active" {
		return nil, ErrSessionClosed
	}

	pr, err := s.Payment.Create(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	s.broadcast(ctx, session, fmt.Sprintf(`{"type":"payment_requested","id":"%s"}`, pr.ID))
	return pr, nil
}

func (s *PaymentService) Confirm(ctx context.Context, id string) error {
	sessionID, err := s.Payment.Confirm(ctx, id)
	if err != nil {
		return err
	}
	if err := s.Session.SetStatus(ctx, sessionID, "payment_requested"); err != nil {
		return err
	}

	session, err := s.Session.FindByID(ctx, sessionID)
	if err != nil {
		return nil
	}
	s.broadcast(ctx, session, fmt.Sprintf(`{"type":"payment_confirmed","sessionId":"%s"}`, sessionID))
	return nil
}

func (s *PaymentService) GetBill(ctx context.Context, sessionID string) (*models.Bill, error) {
	orders, err := s.Order.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var subtotal, vat, total float64
	var included []models.Order
	for _, o := range orders {
		if o.Status == "cancelled" {
			continue
		}
		st, _ := strconv.ParseFloat(o.Subtotal, 64)
		v, _ := strconv.ParseFloat(o.VATAmount, 64)
		t, _ := strconv.ParseFloat(o.Total, 64)
		subtotal += st
		vat += v
		total += t
		included = append(included, o)
	}
	if included == nil {
		included = []models.Order{}
	}

	return &models.Bill{
		SessionID: sessionID,
		Orders:    included,
		Subtotal:  fmt.Sprintf("%.2f", subtotal),
		VATAmount: fmt.Sprintf("%.2f", vat),
		Total:     fmt.Sprintf("%.2f", total),
	}, nil
}

func (s *PaymentService) CloseSession(ctx context.Context, sessionID string) error {
	session, err := s.Session.FindByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	if session.Status != "payment_requested" {
		return ErrPaymentNotConfirmed
	}

	if err := s.Session.Close(ctx, sessionID); err != nil {
		return err
	}
	if err := s.Table.SetStatus(ctx, session.TableID, "ready"); err != nil {
		return err
	}

	s.broadcast(ctx, session, fmt.Sprintf(`{"type":"session_closed","sessionId":"%s"}`, sessionID))
	return nil
}

func (s *PaymentService) broadcast(ctx context.Context, session *models.Session, payload string) {
	if s.Hub == nil {
		return
	}
	table, err := s.Table.FindByID(ctx, session.TableID)
	if err != nil {
		return
	}
	s.Hub.Broadcast("payments:branch:"+table.BranchID, []byte(payload))
	s.Hub.Broadcast("payments:session:"+session.ID, []byte(payload))
}
