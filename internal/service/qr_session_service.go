package service

import (
	"context"
	"errors"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

var (
	ErrQRNotFound      = errors.New("qr not found or expired")
	ErrBranchInactive  = errors.New("branch is not active")
	ErrTableUnavailable = errors.New("table is not available")
)

type QRSessionService struct {
	QR      *repository.QRRepository
	Session *repository.SessionRepository
	Table   *repository.TableRepository
}

type ResolveResult struct {
	Restaurant models.Restaurant
	Branch     models.Branch
	Table      models.Table
	Session    models.Session
}

func (s *QRSessionService) ResolveQRAndCreateSession(ctx context.Context, token string) (*ResolveResult, error) {
	resolved, err := s.QR.FindActiveByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQRNotFound
		}
		return nil, err
	}

	if resolved.Branch.Status != "active" {
		return nil, ErrBranchInactive
	}

	if resolved.Table.Status != "ready" && resolved.Table.Status != "serving" {
		return nil, ErrTableUnavailable
	}

	session, err := s.Session.FindActiveByTable(ctx, resolved.Table.ID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		session, err = s.Session.Create(ctx, resolved.Table.ID)
		if err != nil {
			return nil, err
		}
		if resolved.Table.Status == "ready" {
			if err := s.Table.SetStatus(ctx, resolved.Table.ID, "serving"); err != nil {
				return nil, err
			}
			resolved.Table.Status = "serving"
		}
	}

	return &ResolveResult{
		Restaurant: resolved.Restaurant,
		Branch:     resolved.Branch,
		Table:      resolved.Table,
		Session:    *session,
	}, nil
}
