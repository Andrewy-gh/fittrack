package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	jwksURL     = "https://api.stack-auth.com/api/v1/projects/project_id/.well-known/jwks.json"
	accessToken = "token"
)

func main() {
	ctx := context.Background()

	// 1. Build the cache (HTTP client + background refresh)
	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		log.Fatalf("cannot create JWKS cache: %v", err)
	}

	// 2. Register the URL with refresh policy
	if err := cache.Register(ctx, jwksURL,
		jwk.WithMinInterval(15*time.Minute), // never refresh more often
		jwk.WithMaxInterval(24*time.Hour),   // refresh at least once a day
	); err != nil {
		log.Fatalf("cannot register JWKS URL: %v", err)
	}

	// 3. Wrap the cache as a jwk.Set you can pass around
	cachedSet, err := cache.CachedSet(jwksURL)
	if err != nil {
		log.Fatalf("cannot obtain cached keyset: %v", err)
	}

	// 4. Parse the token using the cached (auto-refreshed) JWKS
	token, err := jwt.ParseString(
		accessToken,
		jwt.WithKeySet(cachedSet),
		jwt.WithValidate(true), // verify exp/nbf/aud etc.
	)
	if err != nil {
		log.Printf("invalid token: %v", err)
		return
	}

	id, ok := token.Subject()
	if !ok {
		log.Println("Invalid user")
		return
	}
	fmt.Printf("Authenticated user ID: %s\n", id)

}
