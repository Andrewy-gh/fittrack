package auth

import (
	"context"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

type JWKS struct {
	keySet jwk.Set
}

func NewJWKS(ctx context.Context, jwksURL string) (*JWKS, error) {
	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, err
	}
	if err = cache.Register(ctx, jwksURL,
		jwk.WithMinInterval(15*time.Minute), // never refresh more often
		jwk.WithMaxInterval(24*time.Hour),   // refresh at least once a day
	); err != nil {
		return nil, err
	}
	cachedSet, err := cache.CachedSet(jwksURL)
	if err != nil {
		return nil, err
	}
	return &JWKS{keySet: cachedSet}, nil
}

func (j *JWKS) KeySet() jwk.Set {
	return j.keySet
}
