package sdk

import (
	"time"
)

// expiryDelta determines how much earlier a token should be considered
// expired than its actual expiration time. It is used to avoid late
// expirations due to client-server time mismatches.
const expiryDelta = 10 * time.Second

// Token represents an OpenFaaS ID token
type Token struct {
	// IDToken is the OIDC access token that authorizes and authenticates
	// the requests.
	IDToken string

	// Expiry is the expiration time of the access token.
	//
	// A zero value means the token never expires.
	Expiry time.Time
}

// Expired reports whether the token is expired, and will start
// to return false 10s before the actual expiration time.
func (t *Token) Expired() bool {
	if t.Expiry.IsZero() {
		return false
	}

	return t.Expiry.Round(0).Add(-expiryDelta).Before(time.Now())
}

type tokenJSON struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`

	TokenType string `json:"token_type"`

	ExpiresIn int `json:"expires_in"`
}

func (t *tokenJSON) expiry() (exp time.Time) {
	if v := t.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}
	return
}
