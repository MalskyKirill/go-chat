package middleware

import (
	"context"
	"encoding/json"
	"go-chat/internal/auth"

	"net/http"
	"strings"
)

type contentKey string

const userIDKey contentKey = "user_id"

func AuthMiddleware(JWTSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractBearerToken(r)
			if tokenString == "" {
				writeError(w, http.StatusUnauthorized, "missing token")
				return
			}

			claims, err := auth.ParseToken(JWTSecret, tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return parts[1]
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
