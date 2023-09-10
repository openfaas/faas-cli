package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Exchange an OIDC ID Token from an IdP for OpenFaaS token
// using the token exchange grant type.
// tokenURL should be the OpenFaaS token endpoint within the internal OIDC service
func ExchangeIDToken(tokenURL, rawIDToken string) (*Token, error) {
	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	v.Set("subject_token_type", "urn:ietf:params:oauth:token-type:id_token")
	v.Set("subject_token", rawIDToken)

	u, _ := url.Parse(tokenURL)

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return nil, fmt.Errorf("cannot fetch token: %v\nResponse: %s", res.Status, body)
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
