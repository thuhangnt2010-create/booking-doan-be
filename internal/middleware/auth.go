package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

type contextKey string

const UserIDKey contextKey = "userId"
const RoleKey contextKey = "role"

func RequireAuth(authService *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Thiếu token đăng nhập")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authService.ParseToken(tokenStr)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token không hợp lệ hoặc hết hạn")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next(w, r.WithContext(ctx))
	}
}

func writeAuthError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": errCode, "message": message})
}
