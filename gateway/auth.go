package gateway

import (
	// "fmt"
	// "log/slog"

	"github.com/golang-jwt/jwt/v5"
)

var jwtParser = jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

func authenticateJWT(tokenString, secret string) (*jwt.Token, error) {
	return jwtParser.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return secret, nil
	})
}