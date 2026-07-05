package handlers

import (
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
)

type PaymentBranchWSHandler struct {
	Hub *realtime.Hub
}

func (h *PaymentBranchWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	branchID := strings.TrimPrefix(r.URL.Path, "/ws/payments/branch/")
	if branchID == "" || branchID == r.URL.Path {
		http.Error(w, "missing branchId", http.StatusBadRequest)
		return
	}
	serveWS(w, r, h.Hub, "payments:branch:"+branchID)
}

type PaymentSessionWSHandler struct {
	Hub *realtime.Hub
}

func (h *PaymentSessionWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/ws/payments/session/")
	if sessionID == "" || sessionID == r.URL.Path {
		http.Error(w, "missing sessionId", http.StatusBadRequest)
		return
	}
	serveWS(w, r, h.Hub, "payments:session:"+sessionID)
}
