package actions

import (
	"beam_payments/actions/firebaseApp"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gobuffalo/buffalo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	id := os.Getenv("GOOGLE_CLIENT_SECRET")
	secret := os.Getenv("GOOGLE_CLIENT_SECRET")

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:3000/test/callback",
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func StartLogin(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("all/ugh.plush.html"))
}

func GoogleLogin(c buffalo.Context) error {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

func GoogleCallback(c buffalo.Context) error {
	code := c.Param("code")
	fmt.Println(code)

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	idToken := token.Extra("id_token").(string)
	fmt.Println(token.AccessToken)
	verifiedToken, err := firebaseApp.FirebaseAuth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		fmt.Println(idToken)
		fmt.Println(err.Error())
		return c.Error(http.StatusUnauthorized, err)
	}

	uid := verifiedToken.UID
	c.Set("UID", uid)
	return c.Render(http.StatusOK, r.HTML("all/ughtwo.plush.html"))
}
