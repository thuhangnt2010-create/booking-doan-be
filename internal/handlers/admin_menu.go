package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

type AdminMenuHandler struct {
	Repo *repository.MenuRepository
	Hub  *realtime.Hub
}

type updateMenuItemRequest struct {
	Price  *string `json:"price"`
	Status *string `json:"status"`
}

func (h *AdminMenuHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/menu-items/")
	if id == "" || id == r.URL.Path {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id món")
		return
	}

	var req updateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}

	branchID, err := h.Repo.UpdateItem(r.Context(), id, req.Price, req.Status)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Món không tồn tại")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	event, _ := json.Marshal(map[string]string{"type": "menu_updated", "itemId": id})
	h.Hub.Broadcast("menu:"+branchID, event)

	w.WriteHeader(http.StatusNoContent)
}
