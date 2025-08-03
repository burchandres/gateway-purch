package gateway

import (
	// "fmt"
	// "log/slog"

	"errors"
	"log/slog"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidPurchClaims = errors.New("invalid purch claims")

type AuthService struct {
	jwtp *jwt.Parser
	config GatewayConfig
}

type PurchClaims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func (p PurchClaims) Validate() error {
	// just check that p.ID is a 32 hex digit of the form 8-4-4-4-12
	splitID := strings.Split(p.ID, "-")
	if len(splitID) != 5 {
		return ErrInvalidPurchClaims
	}
	if len(splitID[0]) != 8 {
		return ErrInvalidPurchClaims
	}
	if len(splitID[1]) != 4 {
		return ErrInvalidPurchClaims
	}
	if len(splitID[2]) != 4 {
		return ErrInvalidPurchClaims
	}
	if len(splitID[3]) != 4 {
		return ErrInvalidPurchClaims
	}
	if len(splitID[4]) != 12 {
		return ErrInvalidPurchClaims
	}
	return nil
}

func NewAuthService() *AuthService {
	return &AuthService{
		config: *ReadConfig(),
		jwtp: jwt.NewParser(
			jwt.WithExpirationRequired(),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		),
	}
}

func (a *AuthService) authenticateJWT(tokenString string) (*PurchClaims, error) {
	token, err := a.jwtp.ParseWithClaims(tokenString, &PurchClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(a.config.Secret), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			slog.Error("invalid tokenString provided", "tokenString", tokenString, "error", err.Error())
			return nil, errors.New("invalid tokenString provided")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			slog.Error("invalid tokenString signature", "tokenString", tokenString, "errror", err.Error())
			return nil, errors.New("invalid tokenString signature")
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			slog.Error("tokenString is expired or not yet valid", "tokenString", tokenString, "error", err.Error())
			return nil, errors.New("tokenString is expired or not yet valid")
		default:
			slog.Error("unable to handle provided tokenString", "tokenString", tokenString, "error", err.Error())
			return nil, errors.New("unable to parse provided tokenString")
		}
	}

	if !token.Valid {
		slog.Error("token is invalid after parsing", "token", token, "tokenString", tokenString)
		return nil, errors.New("token is invalid")
	}

	// extract claims
	claims, ok := token.Claims.(*PurchClaims)
	
	if !ok {
		slog.Error("error extracting custom PurchClaims from token")
		return nil, errors.New("error extracting custom PurchClaims from token")
	}

	return claims, nil
}
