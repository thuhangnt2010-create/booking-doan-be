package handlers

import (
	"net/http"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

type QRImageHandler struct {
	PublicUserURL string
}

func (h *QRImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/qr-images/"), ".png")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	content := h.PublicUserURL + "/?token=" + token
	png, err := qrcode.Encode(content, qrcode.Medium, 320)
	if err != nil {
		http.Error(w, "failed to generate QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}
