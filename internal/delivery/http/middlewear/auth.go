package middlewear

import (
	"context"
	"net/http"
	"refer-mate/internal/usecase"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(authUC *usecase.AuthUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := authUC.ValidateJWT(tokenStr)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) int64 {
	v, _ := r.Context().Value(UserIDKey).(int64)
	return v
}
