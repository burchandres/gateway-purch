package gateway

import (
	// "fmt"
	// "log/slog"

	"github.com/golang-jwt/jwt/v5"
)

type authService struct {
	jwtp *jwt.Parser
}

func (a *authService) authenticateJWT(tokenString, secret string) (*jwt.Token, error) {
	return a.jwtp.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return secret, nil
	})
}
