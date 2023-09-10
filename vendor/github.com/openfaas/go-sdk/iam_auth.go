package sdk

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"
)

// A TokenSource is anything that can return an OIDC ID token that can be exchanged for
// an OpenFaaS token.
type TokenSource interface {
	// Token returns a token or an error.
	Token() (string, error)
}

// TokenAuth bearer token authentication for OpenFaaS deployments with OpenFaaS IAM
// enabled.
type TokenAuth struct {
	// TokenURL represents the OpenFaaS gateways token endpoint URL.
	TokenURL string

	// TokenSource used to get an ID token that can be exchanged for an OpenFaaS ID token.
	TokenSource TokenSource

	lock  sync.Mutex // guards token
	token *Token
}

// Set Authorization Bearer header on request.
// Set validates the token expiry on each call. If it's expired it will exchange
// an ID token from the TokenSource for a new OpenFaaS token.
func (a *TokenAuth) Set(req *http.Request) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.token == nil || a.token.Expired() {
		idToken, err := a.TokenSource.Token()
		if err != nil {
			return err
		}

		token, err := ExchangeIDToken(a.TokenURL, idToken)
		if err != nil {
			return err
		}
		a.token = token
	}

	req.Header.Add("Authorization", "Bearer "+a.token.IDToken)
	return nil
}

// A TokenSource to get ID token by reading a Kubernetes projected service account token
// from /var/secrets/tokens/openfaas-token or the path set by the token_mount_path environment
// variable.
type ServiceAccountTokenSource struct{}

// Token returns a Kubernetes projected service account token read from
// /var/secrets/tokens/openfaas-token or the path set by the token_mount_path
// environment variable.
func (ts *ServiceAccountTokenSource) Token() (string, error) {
	tokenMountPath := getEnv("token_mount_path", "/var/secrets/tokens")
	if len(tokenMountPath) == 0 {
		return "", fmt.Errorf("invalid token_mount_path specified for reading the service account token")
	}

	idTokenPath := path.Join(tokenMountPath, "openfaas-token")
	idToken, err := os.ReadFile(idTokenPath)
	if err != nil {
		return "", fmt.Errorf("unable to load service account token: %s", err)
	}

	return string(idToken), nil
}

func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
