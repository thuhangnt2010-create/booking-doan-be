package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

type SessionExtraHandler struct {
	Service *service.PaymentService
	Auth    *service.AuthService
}

func (h *SessionExtraHandler) SubRoute(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/bill"):
		h.Bill(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/close"):
		h.Close(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *SessionExtraHandler) Bill(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/sessions/"), "/bill")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu session id")
		return
	}

	bill, err := h.Service.GetBill(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bill)
}

func (h *SessionExtraHandler) Close(w http.ResponseWriter, r *http.Request) {
	if !requireAuthInline(w, r, h.Auth) {
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/sessions/"), "/close")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu session id")
		return
	}

	if err := h.Service.CloseSession(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session không tồn tại")
		case errors.Is(err, service.ErrPaymentNotConfirmed):
			writeError(w, http.StatusConflict, "PAYMENT_NOT_CONFIRMED", "Chưa xác nhận yêu cầu thanh toán")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
