package gateway

import (
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
)

var jwtParser = jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

func validateJWT(tokenString, secret string) (bool, error) {
	token, err := jwtParser.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		slog.Debug("error parsing JWT token", "token", tokenString)
		return false, fmt.Errorf("error parsing JWT token %s", tokenString)
	}

	return token.Valid, nil
}