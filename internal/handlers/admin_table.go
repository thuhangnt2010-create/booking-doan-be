package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

type AdminTableHandler struct {
	TableRepo     *repository.TableRepository
	QRRepo        *repository.QRRepository
	PublicUserURL string
}

type createTableBody struct {
	BranchID string `json:"branchId"`
	Area     string `json:"area"`
	Code     string `json:"code"`
}

type qrInfo struct {
	Token      string `json:"token"`
	ImageURL   string `json:"imageUrl"`
	TargetURL  string `json:"targetUrl"`
}

func (h *AdminTableHandler) qrInfoFor(token string) qrInfo {
	if token == "" {
		return qrInfo{}
	}
	return qrInfo{
		Token:     token,
		ImageURL:  "/api/admin/qr-images/" + token + ".png",
		TargetURL: h.PublicUserURL + "/?token=" + token,
	}
}

func (h *AdminTableHandler) Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Create(w, r)
	case http.MethodGet:
		h.List(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *AdminTableHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createTableBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
		return
	}
	if body.BranchID == "" || body.Code == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Thiếu branchId hoặc code")
		return
	}

	table, err := h.TableRepo.Create(r.Context(), body.BranchID, body.Area, body.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống (có thể trùng mã bàn)")
		return
	}

	qr, err := h.QRRepo.CreateForTable(r.Context(), table.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Tạo bàn OK nhưng lỗi sinh QR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"table": table,
		"qr":    h.qrInfoFor(qr.Token),
	})
}

func (h *AdminTableHandler) List(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branchId")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BRANCH_ID", "Thiếu branchId")
		return
	}

	tables, err := h.TableRepo.ListByBranch(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	type tableResponse struct {
		models.TableWithQR
		QR qrInfo `json:"qr"`
	}
	resp := make([]tableResponse, 0, len(tables))
	for _, t := range tables {
		resp = append(resp, tableResponse{TableWithQR: t, QR: h.qrInfoFor(t.ActiveQRToken)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"tables": resp})
}

// SubRoute handles POST /admin/tables/{id}/qr — regenerate QR (old one deactivated).
func (h *AdminTableHandler) SubRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/qr") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	tableID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/tables/"), "/qr")
	if tableID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id bàn")
		return
	}

	qr, err := h.QRRepo.CreateForTable(r.Context(), tableID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"qr": h.qrInfoFor(qr.Token)})
}
