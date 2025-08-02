package gateway

import (
	"strings"
	"net/http"
	"log/slog"
)

func authMiddleware(next http.Handler, secret string, authService *authService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if we're hitting login or register then no authorization needed
		if strings.Contains(r.URL.Path, "/register") || strings.Contains(r.URL.Path, "/token") {
			next.ServeHTTP(w, r)
			return
		}
		// any other hit to purch requires authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			slog.Debug("missing authorization header", "url", r.URL.Path) 
			http.Error(w, "missing authorization headder", http.StatusUnauthorized)
			return
		}
		// verify appropriate authorization header format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			slog.Debug("invalid authorization header", "url", r.URL.Path)
			http.Error(w, "Invalid authorization header format. Expected: Bearer <token>", http.StatusUnauthorized)
			return 
		}
		// retrieve token to authenticate request
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := authService.authenticateJWT(tokenString, secret)
		if err != nil {
			slog.Debug("error validating provided token", "error", err.Error())
			http.Error(w, "error validating provided token", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			slog.Debug("invalid token provided", "token", tokenString)
			http.Error(w, "invalid authorization token.", http.StatusUnauthorized)
			return
		}

	})
}