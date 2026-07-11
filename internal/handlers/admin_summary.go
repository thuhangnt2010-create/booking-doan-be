package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

type AdminSummaryHandler struct {
	OrderRepo   *repository.OrderRepository
	PaymentRepo *repository.PaymentRepository
}

func (h *AdminSummaryHandler) Get(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branchId")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BRANCH_ID", "Thiếu branchId")
		return
	}

	tablesOrdering, err := h.OrderRepo.CountTablesOrdering(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	tablesAwaitingPayment, err := h.PaymentRepo.CountTablesAwaitingPayment(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"tablesOrdering":        tablesOrdering,
		"tablesAwaitingPayment": tablesAwaitingPayment,
	})
}
