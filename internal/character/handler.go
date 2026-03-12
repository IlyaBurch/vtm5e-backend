package character

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vtm5e/backend/internal/auth"
	"github.com/vtm5e/backend/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list characters")
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	userID, _ := auth.UserIDFromContext(r.Context())

	c, err := h.service.Get(r.Context(), id, userID)
	if err != nil {
		switch err.Error() {
		case "character not found":
			response.WriteError(w, http.StatusNotFound, "character not found")
		case "forbidden":
			response.WriteError(w, http.StatusForbidden, "access denied")
		default:
			response.WriteError(w, http.StatusInternalServerError, "failed to get character")
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, c)
}

type characterRequest struct {
	Name string          `json:"name"`
	Clan string          `json:"clan"`
	Data json.RawMessage `json:"data"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req characterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	userID, _ := auth.UserIDFromContext(r.Context())

	c, err := h.service.Create(r.Context(), userID, req.Name, req.Clan, req.Data)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create character")
		return
	}

	response.WriteJSON(w, http.StatusCreated, c)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	var req characterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	userID, _ := auth.UserIDFromContext(r.Context())

	c, err := h.service.Update(r.Context(), id, userID, req.Name, req.Clan, req.Data)
	if err != nil {
		switch err.Error() {
		case "character not found":
			response.WriteError(w, http.StatusNotFound, "character not found")
		case "forbidden":
			response.WriteError(w, http.StatusForbidden, "access denied")
		default:
			response.WriteError(w, http.StatusInternalServerError, "failed to update character")
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	userID, _ := auth.UserIDFromContext(r.Context())

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		switch err.Error() {
		case "character not found":
			response.WriteError(w, http.StatusNotFound, "character not found")
		case "forbidden":
			response.WriteError(w, http.StatusForbidden, "access denied")
		default:
			response.WriteError(w, http.StatusInternalServerError, "failed to delete character")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}
