package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	token        *Token
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

		return ts.token.IDToken, nil
	}

	ts.lock.RUnlock()
	return ts.token.IDToken, nil
}

func obtainClientCredentialsToken(clientID, clientSecret, tokenURL, scope, grantType, audience string) (*Token, error) {

	v := url.Values{}
	v.Set("client_id", clientID)
	v.Set("client_secret", clientSecret)
	v.Set("grant_type", grantType)
	v.Set("scope", scope)

	if len(audience) > 0 {
		v.Set("audience", audience)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "openfaas-go-sdk")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch token: %v", err)
	}

	if code := res.StatusCode; code < 200 || code > 299 {
		return nil, fmt.Errorf("unexpected status code: %v\nResponse: %s", res.Status, body)
	}

	tj := &tokenJSON{}
	if err := json.Unmarshal(body, tj); err != nil {
		return nil, fmt.Errorf("unable to unmarshal token: %s", err)
	}

	return &Token{
		IDToken: tj.AccessToken,
		Expiry:  tj.expiry(),
	}, nil
}
