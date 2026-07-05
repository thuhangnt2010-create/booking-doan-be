package handlers

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type MenuWSHandler struct {
	Hub *realtime.Hub
}

func (h *MenuWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	branchID := strings.TrimPrefix(r.URL.Path, "/ws/menu/")
	if branchID == "" || branchID == r.URL.Path {
		http.Error(w, "missing branchId", http.StatusBadRequest)
		return
	}

	serveWS(w, r, h.Hub, "menu:"+branchID)
}
