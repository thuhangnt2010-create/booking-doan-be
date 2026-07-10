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

type StaffCallHandler struct {
	Service *service.StaffCallService
	Repo    *repository.StaffCallRepository
}

type createStaffCallBody struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"`
}

func (h *StaffCallHandler) Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Create(w, r)
	case http.MethodGet:
		h.List(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *StaffCallHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createStaffCallBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}
	if body.SessionID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SESSION_ID", "Thiếu sessionId")
		return
	}

	call, err := h.Service.Create(r.Context(), body.SessionID, body.Type)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session không tồn tại")
		case errors.Is(err, service.ErrSessionClosed):
			writeError(w, http.StatusConflict, "SESSION_CLOSED", "Session đã đóng")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(call)
}

func (h *StaffCallHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	branchID := q.Get("branchId")
	sessionID := q.Get("sessionId")
	if branchID == "" && sessionID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_QUERY", "Thiếu branchId hoặc sessionId")
		return
	}

	var calls []models.StaffCallRequest
	var err error
	if branchID != "" {
		calls, err = h.Repo.ListByBranch(r.Context(), branchID)
	} else {
		calls, err = h.Repo.ListBySession(r.Context(), sessionID)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if calls == nil {
		calls = []models.StaffCallRequest{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"staffCalls": calls})
}

func (h *StaffCallHandler) SubRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch || !strings.HasSuffix(r.URL.Path, "/status") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/staff-calls/"), "/status")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id")
		return
	}

	var body updateStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}

	if err := h.Service.UpdateStatus(r.Context(), id, body.Status); err != nil {
		writeOrderStatusError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
