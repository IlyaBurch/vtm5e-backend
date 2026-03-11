package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/vtm5e/backend/internal/response"
	"github.com/vtm5e/backend/internal/user"
)

type Handler struct {
	userService *user.Service
	jwtSecret   string
}

func NewHandler(userService *user.Service, jwtSecret string) *Handler {
	return &Handler{userService: userService, jwtSecret: jwtSecret}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string    `json:"token"`
	User  *userInfo `json:"user"`
}

type userInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	u, err := h.userService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		response.WriteError(w, http.StatusConflict, "email already in use")
		return
	}

	token, err := h.generateToken(u.ID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.WriteJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  &userInfo{ID: u.ID, Email: u.Email},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isInvalidCredentials(err) {
			response.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "login failed")
		return
	}

	token, err := h.generateToken(u.ID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.WriteJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  &userInfo{ID: u.ID, Email: u.Email},
	})
}

func (h *Handler) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func isInvalidCredentials(err error) bool {
	return err != nil && err.Error() == "login: invalid credentials"
}
// еуые