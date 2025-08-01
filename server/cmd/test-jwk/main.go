package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	jwksURL     = "https://api.stack-auth.com/api/v1/projects/<project_id>/.well-known/jwks.json"
	accessToken = "token"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Fetch the JWKS from the remote URL
	keySet, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		log.Fatalf("failed to fetch JWKS: %v", err)
	}

	// Parse and verify the JWT token
	token, err := jwt.ParseString(accessToken,
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	)

	if err != nil {
		log.Printf("Failed to verify token: %v\n", err)
		log.Println("Invalid user")
		return
	}

	// If we get here, the token is valid
	fmt.Printf("Token payload: %+v\n", token)
	if id, ok := token.Subject(); ok {
		fmt.Println(id)
	}
}
