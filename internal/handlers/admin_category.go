package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

type AdminCategoryHandler struct {
	Repo *repository.MenuRepository
}

func (h *AdminCategoryHandler) Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.List(w, r)
	case http.MethodPost:
		h.Create(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *AdminCategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branchId")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BRANCH_ID", "Thiếu branchId")
		return
	}
	categories, err := h.Repo.ListCategories(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if categories == nil {
		categories = []models.MenuCategory{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"categories": categories})
}

type createCategoryBody struct {
	BranchID string `json:"branchId"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

func (h *AdminCategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}
	if body.BranchID == "" || body.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Thiếu branchId hoặc name")
		return
	}

	category, err := h.Repo.CreateCategory(r.Context(), body.BranchID, body.Name, body.Position)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}
