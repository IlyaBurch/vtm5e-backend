package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vtm5e/backend/internal/ctxutil"
	"github.com/vtm5e/backend/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())

	u, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	response.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())

	var req struct {
		Username  *string `json:"username"`
		AvatarURL *string `json:"avatarUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.service.UpdateProfile(r.Context(), userID, req.Username, req.AvatarURL)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameTaken):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_TAKEN")
		case errors.Is(err, ErrUsernameInvalid):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_INVALID")
		case errors.Is(err, ErrUsernameTooShort):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_TOO_SHORT")
		case errors.Is(err, ErrUsernameTooLong):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_TOO_LONG")
		default:
			response.WriteError(w, http.StatusInternalServerError, "update failed")
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, u)
}
