package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

type QRHandler struct {
	Service *service.QRSessionService
}

func (h *QRHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/qr/")
	if token == "" || token == r.URL.Path {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Thiếu mã QR")
		return
	}

	result, err := h.Service.ResolveQRAndCreateSession(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrQRNotFound):
			writeError(w, http.StatusNotFound, "QR_NOT_FOUND", "QR không tồn tại hoặc đã hết hiệu lực")
		case errors.Is(err, service.ErrBranchInactive):
			writeError(w, http.StatusForbidden, "BRANCH_INACTIVE", "Chi nhánh hiện đã ngừng hoạt động")
		case errors.Is(err, service.ErrTableUnavailable):
			writeError(w, http.StatusConflict, "TABLE_UNAVAILABLE", "Bàn hiện không sẵn sàng phục vụ")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"restaurant": result.Restaurant,
		"branch":     result.Branch,
		"table":      result.Table,
		"session":    result.Session,
	})
}

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": errCode, "message": message})
}
