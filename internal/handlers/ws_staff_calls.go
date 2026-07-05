package handlers

import (
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
)

type StaffCallBranchWSHandler struct {
	Hub *realtime.Hub
}

func (h *StaffCallBranchWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	branchID := strings.TrimPrefix(r.URL.Path, "/ws/staff-calls/branch/")
	if branchID == "" || branchID == r.URL.Path {
		http.Error(w, "missing branchId", http.StatusBadRequest)
		return
	}
	serveWS(w, r, h.Hub, "staffcalls:branch:"+branchID)
}

type StaffCallSessionWSHandler struct {
	Hub *realtime.Hub
}

func (h *StaffCallSessionWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/ws/staff-calls/session/")
	if sessionID == "" || sessionID == r.URL.Path {
		http.Error(w, "missing sessionId", http.StatusBadRequest)
		return
	}
	serveWS(w, r, h.Hub, "staffcalls:session:"+sessionID)
}
