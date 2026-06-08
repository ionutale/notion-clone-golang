package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
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

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, refreshToken, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	h.setRefreshCookie(w, refreshToken)
	respond(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, refreshToken, err := h.svc.Login(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	h.setRefreshCookie(w, refreshToken)
	respond(w, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "no refresh token")
		return
	}
	resp, newToken, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.clearRefreshCookie(w)
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	h.setRefreshCookie(w, newToken)
	respond(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	h.svc.Logout(r.Context(), userID)
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respond(w, http.StatusOK, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		CurrentPassword string `json:"current_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
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
		respondError(w, code, err.Error())
		return
	}

	respond(w, http.StatusOK, user)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	if err := h.svc.UpdatePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		code := http.StatusInternalServerError
		if err == ErrInvalidCredentials {
			code = http.StatusUnauthorized
		}
		respondError(w, code, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.DeleteAccount(r.Context(), userID, req.Password); err != nil {
		code := http.StatusInternalServerError
		if err == ErrInvalidCredentials {
			code = http.StatusUnauthorized
		}
		respondError(w, code, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}
