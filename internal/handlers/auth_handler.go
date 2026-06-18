package handlers

import (
	"errors"
	"go-chat/internal/dto"
	"go-chat/internal/middleware"
	"go-chat/internal/repositories"
	"go-chat/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.RegisterRequest

	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.authService.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlradyTaken):
			WriteError(w, http.StatusConflict, "email alrady taken")
		case errors.Is(err, service.ErrUsernsmeAlradyTaken):
			WriteError(w, http.StatusConflict, "username already taken")
		default:
			WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.LoginRequest

	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "user not found")
			return
		}

		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, user)
}
