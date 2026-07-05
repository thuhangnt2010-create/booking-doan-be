package handlers

import (
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
)

type OrderBranchWSHandler struct {
	Hub *realtime.Hub
}

func (h *OrderBranchWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	branchID := strings.TrimPrefix(r.URL.Path, "/ws/orders/branch/")
	if branchID == "" || branchID == r.URL.Path {
		http.Error(w, "missing branchId", http.StatusBadRequest)
		return
	}
	serveWS(w, r, h.Hub, "orders:branch:"+branchID)
}

type OrderSessionWSHandler struct {
	Hub *realtime.Hub
}

func (h *OrderSessionWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/ws/orders/session/")
	if sessionID == "" || sessionID == r.URL.Path {
		http.Error(w, "missing sessionId", http.StatusBadRequest)
		return
	}
	serveWS(w, r, h.Hub, "orders:session:"+sessionID)
}

func serveWS(w http.ResponseWriter, r *http.Request, hub *realtime.Hub, topic string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := hub.Register(topic, conn)
	defer func() {
		hub.Unregister(topic, client)
		conn.Close()
	}()
	client.ReadLoop()
}
