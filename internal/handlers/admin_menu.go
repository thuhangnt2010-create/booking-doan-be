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

func (h *AdminMenuHandler) broadcast(branchID, eventType, itemID string) {
	event, _ := json.Marshal(map[string]string{"type": eventType, "itemId": itemID})
	h.Hub.Broadcast("menu:"+branchID, event)
}

type createMenuItemRequest struct {
	CategoryID      string `json:"categoryId"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	Price           string `json:"price"`
	Unit            string `json:"unit"`
	PrepTimeMinutes int    `json:"prepTimeMinutes"`
	Description     string `json:"description"`
	Ingredients     string `json:"ingredients"`
	AllergyInfo     string `json:"allergyInfo"`
	IsPromo         bool   `json:"isPromo"`
	IsBestSeller    bool   `json:"isBestSeller"`
	IsNew           bool   `json:"isNew"`
	Status          string `json:"status"`
}

func (h *AdminMenuHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req createMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}
	if req.CategoryID == "" || req.Name == "" || req.Price == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Thiếu categoryId, name hoặc price")
		return
	}

	id, branchID, err := h.Repo.CreateItem(r.Context(), repository.CreateItemInput{
		CategoryID: req.CategoryID, Code: req.Code, Name: req.Name, Price: req.Price,
		Unit: req.Unit, PrepTimeMinutes: req.PrepTimeMinutes, Description: req.Description,
		Ingredients: req.Ingredients, AllergyInfo: req.AllergyInfo,
		IsPromo: req.IsPromo, IsBestSeller: req.IsBestSeller, IsNew: req.IsNew, Status: req.Status,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống (có thể trùng mã món)")
		return
	}

	detail, err := h.Repo.GetItemDetail(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	h.broadcast(branchID, "menu_updated", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(detail)
}

type updateMenuItemRequest struct {
	CategoryID      *string `json:"categoryId"`
	Code            *string `json:"code"`
	Name            *string `json:"name"`
	Price           *string `json:"price"`
	Unit            *string `json:"unit"`
	PrepTimeMinutes *int    `json:"prepTimeMinutes"`
	Description     *string `json:"description"`
	Ingredients     *string `json:"ingredients"`
	AllergyInfo     *string `json:"allergyInfo"`
	IsPromo         *bool   `json:"isPromo"`
	IsBestSeller    *bool   `json:"isBestSeller"`
	IsNew           *bool   `json:"isNew"`
	Status          *string `json:"status"`
}

func (h *AdminMenuHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	var req updateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}

	branchID, err := h.Repo.UpdateItemFull(r.Context(), id, repository.UpdateItemInput{
		CategoryID: req.CategoryID, Code: req.Code, Name: req.Name, Price: req.Price,
		Unit: req.Unit, PrepTimeMinutes: req.PrepTimeMinutes, Description: req.Description,
		Ingredients: req.Ingredients, AllergyInfo: req.AllergyInfo,
		IsPromo: req.IsPromo, IsBestSeller: req.IsBestSeller, IsNew: req.IsNew, Status: req.Status,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Món không tồn tại")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	h.broadcast(branchID, "menu_updated", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminMenuHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	branchID, err := h.Repo.DeleteItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Món không tồn tại")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	h.broadcast(branchID, "menu_deleted", id)
	w.WriteHeader(http.StatusNoContent)
}

// ItemSubRoute handles PATCH/DELETE on "/admin/menu-items/{id}".
func (h *AdminMenuHandler) ItemSubRoute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/menu-items/")
	if id == "" || id == r.URL.Path {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id món")
		return
	}
	switch r.Method {
	case http.MethodPatch:
		h.Update(w, r, id)
	case http.MethodDelete:
		h.Delete(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
