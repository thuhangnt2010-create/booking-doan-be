package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

type PaymentHandler struct {
	Service *service.PaymentService
}

type createPaymentRequestBody struct {
	SessionID string `json:"sessionId"`
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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
