package firebaseApp

import (
	"context"
	"errors"
	"strings"

	"github.com/gobuffalo/buffalo"
)

func VerifyTokenAndReturnUID(c buffalo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", errors.New("invalid token format")
	}

	token, err := FirebaseAuth.VerifyIDToken(context.Background(), tokenString)
	if err != nil {
		return "", errors.New("failed to verify token")
	}

	return token.UID, nil
}
