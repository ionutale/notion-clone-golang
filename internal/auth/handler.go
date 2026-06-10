package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ionutale/notion-clone-golang/internal/httputil"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	svc    *Service
	secure bool
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, secure: true}
}

func NewHandlerDev(svc *Service) *Handler {
	return &Handler{svc: svc, secure: false}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
}

func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Post("/auth/logout", h.Logout)
	r.Get("/auth/me", h.Me)
	r.Patch("/auth/me", h.UpdateProfile)
	r.Patch("/auth/me/password", h.UpdatePassword)
	r.Delete("/auth/me", h.DeleteAccount)
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		MaxAge:   -1,
	})
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		httputil.Error(w, http.StatusBadRequest, "valid email is required")
		return
	}
	if len(req.Password) < 8 {
		httputil.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	resp, refreshToken, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err.Error())
		return
	}
	h.setRefreshCookie(w, refreshToken, h.secure)
	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		httputil.Error(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "password is required")
		return
	}

	resp, refreshToken, err := h.svc.Login(r.Context(), req)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	h.setRefreshCookie(w, refreshToken, h.secure)
	httputil.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "no refresh token")
		return
	}
	resp, newToken, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.clearRefreshCookie(w, h.secure)
		httputil.Error(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	h.setRefreshCookie(w, newToken, h.secure)
	httputil.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	if err := h.svc.Logout(r.Context(), userID); err != nil {
		slog.Error("logout failed", "error", err)
	}
	h.clearRefreshCookie(w, h.secure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "user not found")
		return
	}
	httputil.JSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		CurrentPassword string `json:"current_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email != "" && !strings.Contains(req.Email, "@") {
		httputil.Error(w, http.StatusBadRequest, "valid email is required")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), userID, req.Name, req.Email, req.CurrentPassword)
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case ErrInvalidCredentials:
			code = http.StatusUnauthorized
		case ErrEmailTaken:
			code = http.StatusConflict
		}
		httputil.Error(w, code, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		httputil.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if req.CurrentPassword == "" {
		httputil.Error(w, http.StatusBadRequest, "current password is required")
		return
	}

	if err := h.svc.UpdatePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		code := http.StatusInternalServerError
		if err == ErrInvalidCredentials {
			code = http.StatusUnauthorized
		}
		httputil.Error(w, code, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "password is required")
		return
	}

	if err := h.svc.DeleteAccount(r.Context(), userID, req.Password); err != nil {
		code := http.StatusInternalServerError
		if err == ErrInvalidCredentials {
			code = http.StatusUnauthorized
		}
		httputil.Error(w, code, err.Error())
		return
	}

	h.clearRefreshCookie(w, h.secure)
	w.WriteHeader(http.StatusNoContent)
}


