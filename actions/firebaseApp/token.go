package firebaseApp

import (
	"context"
	"errors"
	"strings"

	"github.com/gobuffalo/buffalo"
)

func VerifyTokenAndReturnUID(c buffalo.Context) (string, error) {
	idToken := c.Request().FormValue("idToken")
	if idToken == "" {
		return "", errors.New("missing idToken in form data")
	}

	token, err := FirebaseAuth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", errors.New("failed to verify token")
	}

	return token.UID, nil
}

func VerifyTokenAndReturnUIDBearer(c buffalo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid Authorization header format")
	}

	idToken := parts[1]

	token, err := FirebaseAuth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", errors.New("failed to verify token")
	}

	return token.UID, nil
}
