package sdk

import (
	"fmt"
	"net/http"
)

const DefaultNamespace = "openfaas-fn"

func (c *Client) InvokeFunction(name, namespace string, async bool, auth bool, req *http.Request) (*http.Response, error) {
	fnEndpoint := "/function"
	if async {
		fnEndpoint = "/async-function"
	}

	if len(namespace) == 0 {
		namespace = DefaultNamespace
	}

	req.URL.Scheme = c.GatewayURL.Scheme
	req.URL.Host = c.GatewayURL.Host
	req.URL.Path = fmt.Sprintf("%s/%s.%s", fnEndpoint, name, namespace)

	if auth && c.FunctionTokenSource != nil {
		idToken, err := c.FunctionTokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to get function access token: %w", err)
		}

		tokenURL := fmt.Sprintf("%s/oauth/token", c.GatewayURL.String())
		scope := []string{"function"}
		audience := []string{fmt.Sprintf("%s:%s", namespace, name)}

		var bearer string
		if c.fnTokenCache != nil {
			// Function access tokens are cached as long as the token is valid
			// to prevent having to do a token exchange each time the function is invoked.
			cacheKey := fmt.Sprintf("%s.%s", name, namespace)
			token, ok := c.fnTokenCache.Get(cacheKey)
			if !ok {
				token, err = ExchangeIDToken(tokenURL, idToken, WithScope(scope), WithAudience(audience))
				if err != nil {
					return nil, fmt.Errorf("failed to get function access token: %w", err)
				}

				c.fnTokenCache.Set(cacheKey, token)
			}

			bearer = token.IDToken
		} else {
			token, err := ExchangeIDToken(tokenURL, idToken, WithScope(scope), WithAudience(audience))
			if err != nil {
				return nil, fmt.Errorf("failed to get function access token: %w", err)
			}

			bearer = token.IDToken
		}

		req.Header.Add("Authorization", "Bearer "+bearer)
	}

	return c.do(req)
}
