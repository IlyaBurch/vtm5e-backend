package morkborg

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vtm5e/backend/internal/ctxutil"
	"github.com/vtm5e/backend/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())

	chars, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list characters")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"characters": chars})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.service.Create(r.Context(), userID, &req)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]any{"character": c})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	c, err := h.service.Get(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.WriteError(w, http.StatusForbidden, "FORBIDDEN")
			return
		}
		response.WriteError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"character": c})
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.service.Patch(r.Context(), id, userID, &req)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.WriteError(w, http.StatusForbidden, "FORBIDDEN")
			return
		}
		response.WriteError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"character": c})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := ctxutil.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrForbidden) {
			response.WriteError(w, http.StatusForbidden, "FORBIDDEN")
			return
		}
		response.WriteError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
