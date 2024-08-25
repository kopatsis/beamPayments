package firebaseApp

import (
	"context"
	"encoding/base64"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

func init() {
	firebaseConfigBase64 := os.Getenv("FIREBASE_CONFIG_BASE64")
	if firebaseConfigBase64 == "" {
		log.Fatal("FIREBASE_CONFIG_BASE64 environment variable is not set.")
	}

	configJSON, err := base64.StdEncoding.DecodeString(firebaseConfigBase64)
	if err != nil {
		log.Fatalf("Error decoding FIREBASE_CONFIG_BASE64: %v", err)
	}

	sa := option.WithCredentialsJSON(configJSON)
	app, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v\n", err)
	}

	FirebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v\n", err)
	}
}
