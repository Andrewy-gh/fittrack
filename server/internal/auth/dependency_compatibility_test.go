package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

// Sign wire claims independently of jwx so parsing changes cannot be hidden by
// matching changes to the library's token builder or serializer.
func TestJWKSCache_VerifiesIndependentWireTokens(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	public, err := jwk.Import(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := public.Set(jwk.KeyIDKey, "wire-key"); err != nil {
		t.Fatal(err)
	}
	keys := jwk.NewSet()
	if err := keys.AddKey(public); err != nil {
		t.Fatal(err)
	}
	cache := &JWKSCache{keySet: keys}
	now := time.Now().Unix()
	for _, tc := range []struct {
		name, claims   string
		tamper, reject bool
	}{
		{"integer dates", fmt.Sprintf(`{"sub":"wire-user","iat":%d,"exp":%d}`, now-30, now+300), false, false},
		{"decimal dates", fmt.Sprintf(`{"sub":"wire-user","iat":%d.125,"exp":%d.875}`, now-30, now+300), false, false},
		{"negative issued at", fmt.Sprintf(`{"sub":"wire-user","iat":-1,"exp":%d}`, now+300), false, false},
		{"expired", fmt.Sprintf(`{"sub":"wire-user","exp":%d}`, now-300), false, true},
		{"future issued at", fmt.Sprintf(`{"sub":"wire-user","iat":%d,"exp":%d}`, now+300, now+600), false, true},
		{"invalid signature", fmt.Sprintf(`{"sub":"wire-user","exp":%d}`, now+300), true, true},
		{"missing subject", fmt.Sprintf(`{"exp":%d}`, now+300), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encode := base64.RawURLEncoding.EncodeToString
			message := encode([]byte(`{"alg":"RS256","kid":"wire-key"}`)) + "." + encode([]byte(tc.claims))
			digest := sha256.Sum256([]byte(message))
			signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
			if err != nil {
				t.Fatal(err)
			}
			if tc.tamper {
				signature[0] ^= 1
			}
			userID, err := cache.GetUserIDFromToken(message + "." + encode(signature))
			if tc.reject {
				if err == nil {
					t.Fatal("invalid token accepted")
				}
				return
			}
			if err != nil || userID != "wire-user" {
				t.Fatalf("user=%q err=%v", userID, err)
			}
		})
	}
}
