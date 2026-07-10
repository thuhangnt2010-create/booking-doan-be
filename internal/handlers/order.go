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

type OrderHandler struct {
	Service   *service.OrderService
	OrderRepo *repository.OrderRepository
}

type createOrderItemBody struct {
	ItemID    string   `json:"itemId"`
	Qty       int      `json:"qty"`
	Note      string   `json:"note"`
	OptionIDs []string `json:"optionIds"`
}

type createOrderBody struct {
	SessionID       string                 `json:"sessionId"`
	ClientRequestID string                 `json:"clientRequestId"`
	Items           []createOrderItemBody  `json:"items"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createOrderBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}
	if body.SessionID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SESSION_ID", "Thiếu sessionId")
		return
	}

	req := service.CreateOrderRequest{SessionID: body.SessionID, ClientRequestID: body.ClientRequestID}
	for _, it := range body.Items {
		req.Items = append(req.Items, service.OrderItemRequest{
			ItemID: it.ItemID, Qty: it.Qty, Note: it.Note, OptionIDs: it.OptionIDs,
		})
	}

	order, err := h.Service.CreateOrder(r.Context(), req)
	if err != nil {
		var valErr *service.OrderValidationError
		switch {
		case errors.As(err, &valErr):
			writeError(w, http.StatusBadRequest, valErr.Code, valErr.Message)
		case errors.Is(err, service.ErrSessionNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session không tồn tại")
		case errors.Is(err, service.ErrSessionClosed):
			writeError(w, http.StatusConflict, "SESSION_CLOSED", "Session đã đóng hoặc đang chờ thanh toán")
		case errors.Is(err, service.ErrEmptyOrder):
			writeError(w, http.StatusBadRequest, "EMPTY_ORDER", "Order phải có ít nhất 1 món")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	branchID := q.Get("branchId")
	sessionID := q.Get("sessionId")
	if branchID == "" && sessionID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_QUERY", "Thiếu branchId hoặc sessionId")
		return
	}

	var orders []models.Order
	var err error
	if branchID != "" {
		orders, err = h.OrderRepo.ListByBranch(r.Context(), branchID)
	} else {
		orders, err = h.OrderRepo.ListBySession(r.Context(), sessionID)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

type updateStatusBody struct {
	Status string `json:"status"`
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/orders/"), "/status")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id order")
		return
	}

	var body updateStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}

	if err := h.Service.UpdateOrderStatus(r.Context(), id, body.Status); err != nil {
		writeOrderStatusError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *OrderHandler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/order-items/"), "/status")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id món trong order")
		return
	}

	var body updateStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}

	if err := h.Service.UpdateOrderItemStatus(r.Context(), id, body.Status); err != nil {
		writeOrderStatusError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Root dispatches exact "/orders" by method: POST creates, GET lists by ?sessionId=.
func (h *OrderHandler) Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Create(w, r)
	case http.MethodGet:
		h.List(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// SubRoute handles "/orders/{id}/status" (PATCH).
func (h *OrderHandler) SubRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/status") {
		h.UpdateStatus(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// ItemSubRoute handles "/order-items/{id}/status" (PATCH).
func (h *OrderHandler) ItemSubRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/status") {
		h.UpdateItemStatus(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func writeOrderStatusError(w http.ResponseWriter, err error) {
	var valErr *service.OrderValidationError
	switch {
	case errors.As(err, &valErr):
		writeError(w, http.StatusBadRequest, valErr.Code, valErr.Message)
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
	}
}
