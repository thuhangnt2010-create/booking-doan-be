package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

var validStaffCallStatuses = map[string]bool{
	"sent": true, "received": true, "processing": true, "done": true,
}

type StaffCallService struct {
	Session   *repository.SessionRepository
	Table     *repository.TableRepository
	StaffCall *repository.StaffCallRepository
	Hub       *realtime.Hub
}

func (s *StaffCallService) Create(ctx context.Context, sessionID, callType string) (*models.StaffCallRequest, error) {
	session, err := s.Session.FindByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if session.Status == "closed" {
		return nil, ErrSessionClosed
	}
	if callType == "" {
		callType = "other"
	}

	call, err := s.StaffCall.Create(ctx, sessionID, callType)
	if err != nil {
		return nil, err
	}

	payload := fmt.Sprintf(`{"type":"staff_call_created","id":"%s","callType":"%s"}`, call.ID, call.Type)
	s.broadcast(ctx, session, payload)

	return call, nil
}

func (s *StaffCallService) UpdateStatus(ctx context.Context, id, status string) error {
	if !validStaffCallStatuses[status] {
		return &OrderValidationError{Code: "INVALID_STATUS", Message: "Trạng thái không hợp lệ"}
	}
	sessionID, err := s.StaffCall.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	session, err := s.Session.FindByID(ctx, sessionID)
	if err != nil {
		return nil
	}
	payload := fmt.Sprintf(`{"type":"staff_call_status","id":"%s","status":"%s"}`, id, status)
	s.broadcast(ctx, session, payload)
	return nil
}

func (s *StaffCallService) broadcast(ctx context.Context, session *models.Session, payload string) {
	if s.Hub == nil {
		return
	}
	table, err := s.Table.FindByID(ctx, session.TableID)
	if err != nil {
		return
	}
	s.Hub.Broadcast("staffcalls:branch:"+table.BranchID, []byte(payload))
	s.Hub.Broadcast("staffcalls:session:"+session.ID, []byte(payload))
}
