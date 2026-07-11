package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

type PaymentHandler struct {
	Service *service.PaymentService
	Repo    *repository.PaymentRepository
	Auth    *service.AuthService
}

type createPaymentRequestBody struct {
	SessionID string `json:"sessionId"`
}

func (h *PaymentHandler) Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Create(w, r)
	case http.MethodGet:
		h.List(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	if !requireAuthInline(w, r, h.Auth) {
		return
	}
	branchID := r.URL.Query().Get("branchId")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BRANCH_ID", "Thiếu branchId")
		return
	}
	requests, err := h.Repo.ListByBranch(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if requests == nil {
		requests = []models.PaymentRequest{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"paymentRequests": requests})
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createPaymentRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}
	if body.SessionID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SESSION_ID", "Thiếu sessionId")
		return
	}

	pr, err := h.Service.RequestPayment(r.Context(), body.SessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session không tồn tại")
		case errors.Is(err, service.ErrSessionClosed):
			writeError(w, http.StatusConflict, "SESSION_NOT_ACTIVE", "Session không ở trạng thái đang phục vụ")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pr)
}

func (h *PaymentHandler) SubRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch || !strings.HasSuffix(r.URL.Path, "/confirm") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if !requireAuthInline(w, r, h.Auth) {
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/payment-requests/"), "/confirm")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id")
		return
	}

	if err := h.Service.Confirm(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
