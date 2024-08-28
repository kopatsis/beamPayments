package firebaseApp

import (
	"context"
	"errors"

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
