package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

type MenuHandler struct {
	Repo *repository.MenuRepository
}

func (h *MenuHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	branchID := q.Get("branchId")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BRANCH_ID", "Thiếu branchId")
		return
	}

	filter := repository.MenuFilter{
		BranchID:   branchID,
		Search:     q.Get("search"),
		CategoryID: q.Get("category"),
		MinPrice:   q.Get("minPrice"),
		MaxPrice:   q.Get("maxPrice"),
		Promo:      q.Get("promo") == "true",
		BestSeller: q.Get("bestSeller") == "true",
		IsNew:      q.Get("isNew") == "true",
		Sort:       strings.ToLower(q.Get("sort")),
	}

	items, err := h.Repo.ListItems(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if items == nil {
		items = []models.MenuItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"items": items})
}

func (h *MenuHandler) Detail(w http.ResponseWriter, r *http.Request) {
	itemID := strings.TrimPrefix(r.URL.Path, "/menu-items/")
	if itemID == "" || itemID == r.URL.Path {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id món")
		return
	}

	detail, err := h.Repo.GetItemDetail(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Món không tồn tại")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if detail.Options == nil {
		detail.Options = []models.MenuItemOption{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}
