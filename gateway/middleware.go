package gateway

import (
	"log/slog"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler, authService AuthService) http.Handler {
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
		if _, err := authService.authenticateJWT(tokenString); err != nil {
			slog.Debug("error authenticating provided tokenString", "tokenString", tokenString, "error", err.Error())
			http.Error(w, "error authenticating provided tokenString", http.StatusUnauthorized)
			return
		}



		// ctx := context.WithValue(r.Context(), "user", user)

		// next.ServeHTTP(w, r.WithContext(ctx))
	})
}