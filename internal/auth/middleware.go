package auth

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vtm5e/backend/internal/ctxutil"
	"github.com/vtm5e/backend/internal/response"
)

func Middleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.WriteError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				response.WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				response.WriteError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			ctx := ctxutil.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
