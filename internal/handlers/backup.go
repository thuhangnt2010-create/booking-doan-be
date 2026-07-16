package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/middleware"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

type BackupHandler struct {
	Repo   *repository.BackupRepository
	Backup *service.BackupService
	OAuth  *service.BackupOAuthService
	// RedirectBase is the public API base URL (e.g. http://<host>/api) used to
	// build the OAuth callback URL registered in Google Cloud Console.
	RedirectBase string
}

// requireBackupAdmin restricts backup config/trigger to the "admin" role —
// staff/kitchen accounts must not be able to read credentials or run backups.
func requireBackupAdmin(w http.ResponseWriter, r *http.Request) bool {
	role, _ := r.Context().Value(middleware.RoleKey).(string)
	if role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Chỉ Quản lý mới có quyền sao lưu")
		return false
	}
	return true
}

func (h *BackupHandler) Settings(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		s, err := h.Repo.GetSettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	case http.MethodPut:
		var body models.BackupSettings
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "Body không hợp lệ")
			return
		}
		userID, _ := r.Context().Value(middleware.UserIDKey).(string)
		s, err := h.Repo.UpdateSettings(r.Context(), body, &userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *BackupHandler) RunNow(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	busy, err := h.Repo.HasRunningBackup(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if busy {
		writeError(w, http.StatusConflict, "BACKUP_RUNNING", "Đang có 1 lượt sao lưu chạy, vui lòng đợi")
		return
	}

	settings, err := h.Repo.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if settings.GdriveAccountEmail == "" {
		writeError(w, http.StatusBadRequest, "GDRIVE_NOT_CONNECTED", "Chưa kết nối Google Drive")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	run, err := h.Repo.CreateRun(r.Context(), "manual", settings.ScopeDB, settings.ScopeUploads, settings.ScopeConfigs, &userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}

	go h.Backup.RunBackup(context.Background(), run.ID, settings)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(run)
}

func (h *BackupHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	limit := 7
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	runs, err := h.Repo.ListRuns(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if runs == nil {
		runs = []models.BackupRun{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"runs": runs})
}

func (h *BackupHandler) RunSubRoute(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/backup/runs/")
	if id == "" || id == r.URL.Path {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Thiếu id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		run, err := h.Repo.GetRun(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy")
				return
			}
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	case http.MethodDelete:
		if err := h.Repo.DeleteRun(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy")
				return
			}
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *BackupHandler) RcloneStatus(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	versionOut, err := exec.CommandContext(r.Context(), "rclone", "version").CombinedOutput()
	installed := err == nil
	version := ""
	if lines := strings.SplitN(string(versionOut), "\n", 2); len(lines) > 0 {
		version = strings.TrimSpace(lines[0])
	}
	remotesOut, _ := exec.CommandContext(r.Context(), "rclone", "listremotes").Output()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"installed": installed,
		"version":   version,
		"remotes":   strings.Fields(string(remotesOut)),
	})
}

type gdriveCredsBody struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func (h *BackupHandler) GdriveCreds(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body gdriveCredsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClientID == "" || body.ClientSecret == "" {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Thiếu Client ID hoặc Client Secret")
		return
	}
	if err := h.Repo.UpdateGdriveCreds(r.Context(), body.ClientID, body.ClientSecret); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BackupHandler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	settings, err := h.Repo.GetSettings(r.Context())
	if err != nil || settings.GdriveClientID == "" || settings.GdriveClientSecret == "" {
		writeError(w, http.StatusBadRequest, "MISSING_CREDS", "Chưa lưu Client ID / Client Secret")
		return
	}
	redirectURL := h.RedirectBase + "/public/backup/oauth/callback"
	authURL, err := h.OAuth.BuildAuthURL(r.Context(), settings.GdriveClientID, settings.GdriveClientSecret, redirectURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"authUrl": authURL, "redirectUrl": redirectURL})
}

// OAuthCallback is registered PUBLIC (no Bearer token — Google's redirect
// never carries one). CSRF protection comes from the one-time state param.
func (h *BackupHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code, state := q.Get("code"), q.Get("state")

	fail := func(msg string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html><body><script>
window.opener && window.opener.postMessage({type:"backup-oauth-result", ok:false, msg:%q}, "*");
window.close();
</script>Đăng nhập thất bại: %s</body></html>`, msg, msg)
	}

	if code == "" || state == "" {
		fail("Thiếu code hoặc state")
		return
	}
	if !h.OAuth.ValidateState(r.Context(), state) {
		fail("State không hợp lệ hoặc đã hết hạn")
		return
	}

	settings, err := h.Repo.GetSettings(r.Context())
	if err != nil {
		fail("Lỗi hệ thống")
		return
	}

	redirectURL := h.RedirectBase + "/public/backup/oauth/callback"
	email, err := h.OAuth.ExchangeAndStore(r.Context(), code, settings.GdriveClientID, settings.GdriveClientSecret, redirectURL, settings.RcloneRemote)
	if err != nil {
		fail(err.Error())
		return
	}
	if err := h.Repo.UpdateGdriveAccount(r.Context(), email); err != nil {
		fail("Lưu tài khoản thất bại")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><body><script>
window.opener && window.opener.postMessage({type:"backup-oauth-result", ok:true, msg:%q}, "*");
window.close();
</script>Kết nối thành công: %s</body></html>`, email, email)
}

func (h *BackupHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if !requireBackupAdmin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	settings, err := h.Repo.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống")
		return
	}
	if err := h.OAuth.Disconnect(r.Context(), settings.RcloneRemote); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
