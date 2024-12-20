package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Exchange an OIDC ID Token from an IdP for OpenFaaS token
// using the token exchange grant type.
// tokenURL should be the OpenFaaS token endpoint within the internal OIDC service
func ExchangeIDToken(tokenURL, rawIDToken string, options ...ExchangeOption) (*Token, error) {
	c := &ExchangeConfig{
		Client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	v.Set("subject_token_type", "urn:ietf:params:oauth:token-type:id_token")
	v.Set("subject_token", rawIDToken)

	for _, aud := range c.Audience {
		v.Add("audience", aud)
	}

	if len(c.Scope) > 0 {
		v.Set("scope", strings.Join(c.Scope, " "))
	}

	u, _ := url.Parse(tokenURL)

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "openfaas-go-sdk")

	if os.Getenv("FAAS_DEBUG") == "1" {
		dump, err := dumpRequest(req)
		if err != nil {
			return nil, err
		}

		fmt.Println(dump)
	}

	res, err := c.Client.Do(req)
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

	if res.StatusCode == http.StatusBadRequest {
		authErr := &OAuthError{}
		if err := json.Unmarshal(body, authErr); err == nil {
			return nil, authErr
		}
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
		Scope:   tj.scope(),
	}, nil
}

type ExchangeConfig struct {
	Audience []string
	Scope    []string
	Client   *http.Client
}

// ExchangeOption is used to implement functional-style options that modify the
// config setting for the OpenFaaS token exchange.
type ExchangeOption func(*ExchangeConfig)

// WithAudience is an option to configure the audience requested
// in the token exchange.
func WithAudience(audience []string) ExchangeOption {
	return func(c *ExchangeConfig) {
		c.Audience = audience
	}
}

// WithScope is an option to configure the scope requested
// in the token exchange.
func WithScope(scope []string) ExchangeOption {
	return func(c *ExchangeConfig) {
		c.Scope = scope
	}
}

// WithHttpClient is an option to configure the http client
// used to make the token exchange request.
func WithHttpClient(client *http.Client) ExchangeOption {
	return func(c *ExchangeConfig) {
		c.Client = client
	}
}
