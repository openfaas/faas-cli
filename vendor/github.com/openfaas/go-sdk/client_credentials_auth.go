package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ClientCredentialsAuth struct {
	tokenSource TokenSource
}

func NewClientCredentialsAuth(ts TokenSource) *ClientCredentialsAuth {
	return &ClientCredentialsAuth{
		tokenSource: ts,
	}
}

func (cca *ClientCredentialsAuth) Set(req *http.Request) error {
	token, err := cca.tokenSource.Token()
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return nil
}

// ClientCredentialsTokenSource can be used to obtain
// an access token using the client credentials grant type.
// Tested with Keycloak's token endpoint, additional changes may
// be required for additional OIDC token endpoints.
type ClientCredentialsTokenSource struct {
	clientID     string
	clientSecret string
	tokenURL     string
	scope        string
	grantType    string
	audience     string
	token        *ClientCredentialsToken
	lock         sync.RWMutex
}

func NewClientCredentialsTokenSource(clientID, clientSecret, tokenURL, scope, grantType, audience string) TokenSource {
	return &ClientCredentialsTokenSource{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		scope:        scope,
		grantType:    grantType,
		audience:     audience,
	}
}

func (ts *ClientCredentialsTokenSource) Token() (string, error) {
	ts.lock.RLock()
	expired := ts.token == nil || ts.token.Expired()

	if expired {
		ts.lock.RUnlock()

		token, err := obtainClientCredentialsToken(ts.clientID, ts.clientSecret, ts.tokenURL, ts.scope, ts.grantType, ts.audience)
		if err != nil {
			return "", err
		}

		ts.lock.Lock()
		ts.token = token
		ts.lock.Unlock()

		return token.AccessToken, nil
	}

	ts.lock.RUnlock()
	return ts.token.AccessToken, nil
}

func obtainClientCredentialsToken(clientID, clientSecret, tokenURL, scope, grantType, audience string) (*ClientCredentialsToken, error) {

	reqBody := url.Values{}
	reqBody.Set("client_id", clientID)
	reqBody.Set("client_secret", clientSecret)
	reqBody.Set("grant_type", grantType)
	reqBody.Set("scope", scope)

	if len(audience) > 0 {
		reqBody.Set("audience", audience)
	}

	buffer := []byte(reqBody.Encode())

	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewBuffer(buffer))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d, body: %s", res.StatusCode, string(body))
	}

	token := &ClientCredentialsToken{}
	if err := json.Unmarshal(body, token); err != nil {
		return nil, fmt.Errorf("unable to unmarshal token: %s", err)
	}

	token.ObtainedAt = time.Now()

	return token, nil
}

// ClientCredentialsToken represents an access_token
// obtained through the client credentials grant type.
// This token is not associated with a human user.
type ClientCredentialsToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ObtainedAt  time.Time
}

// Expired returns true if the token is expired
// or if the expiry time is not known.
// The token will always expire 10s early to avoid
// clock skew.
func (t *ClientCredentialsToken) Expired() bool {
	if t.ExpiresIn == 0 {
		return false
	}
	expiry := t.ObtainedAt.Add(time.Duration(t.ExpiresIn) * time.Second).Add(-expiryDelta)

	return expiry.Before(time.Now())
}
