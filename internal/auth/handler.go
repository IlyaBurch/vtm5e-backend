package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/vtm5e/backend/internal/response"
	"github.com/vtm5e/backend/internal/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Handler struct {
	userService *user.Service
	jwtSecret   string
	oauthConfig *oauth2.Config
	frontendURL string
}

func NewHandler(userService *user.Service, jwtSecret string, oauthConfig *oauth2.Config, frontendURL string) *Handler {
	return &Handler{
		userService: userService,
		jwtSecret:   jwtSecret,
		oauthConfig: oauthConfig,
		frontendURL: frontendURL,
	}
}

func NewOAuthConfig(clientID, clientSecret, callbackURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string    `json:"token"`
	User  *userInfo `json:"user"`
}

type userInfo struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Username  *string `json:"username"`
	AvatarURL *string `json:"avatarUrl"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		response.WriteError(w, http.StatusBadRequest, "email, password and username are required")
		return
	}

	u, err := h.userService.Register(r.Context(), req.Email, req.Password, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUsernameTaken):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_TAKEN")
		case errors.Is(err, user.ErrUsernameInvalid):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_INVALID")
		case errors.Is(err, user.ErrUsernameTooShort):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_TOO_SHORT")
		case errors.Is(err, user.ErrUsernameTooLong):
			response.WriteError(w, http.StatusUnprocessableEntity, "USERNAME_TOO_LONG")
		default:
			response.WriteError(w, http.StatusConflict, "email already in use")
		}
		return
	}

	token, err := h.generateToken(u.ID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.WriteJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  toUserInfo(u),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
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
		User:  toUserInfo(u),
	})
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig.ClientID == "" {
		response.WriteError(w, http.StatusNotImplemented, "Google OAuth not configured")
		return
	}

	state, err := generateState()
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	failURL := h.frontendURL + "/auth?error=google_failed"

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Redirect(w, r, failURL, http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, failURL, http.StatusTemporaryRedirect)
		return
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, failURL, http.StatusTemporaryRedirect)
		return
	}

	googleUser, err := fetchGoogleUser(token.AccessToken)
	if err != nil {
		http.Redirect(w, r, failURL, http.StatusTemporaryRedirect)
		return
	}

	var avatarURL *string
	if googleUser.Picture != "" {
		avatarURL = &googleUser.Picture
	}

	u, err := h.userService.UpsertGoogleUser(r.Context(), googleUser.ID, googleUser.Email, googleUser.Name, avatarURL)
	if err != nil {
		http.Redirect(w, r, failURL, http.StatusTemporaryRedirect)
		return
	}

	jwtToken, err := h.generateToken(u.ID)
	if err != nil {
		http.Redirect(w, r, failURL, http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		MaxAge:   30 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Path:     "/",
	})

	http.Redirect(w, r, h.frontendURL+"/characters", http.StatusTemporaryRedirect)
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleUser(accessToken string) (*googleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("fetch google user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read google user response: %w", err)
	}

	var u googleUserInfo
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("parse google user: %w", err)
	}
	return &u, nil
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

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func toUserInfo(u *user.User) *userInfo {
	return &userInfo{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
	}
}

func isInvalidCredentials(err error) bool {
	return err != nil && err.Error() == "login: invalid credentials"
}
