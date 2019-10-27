package proxy

import (
	"context"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"
)

// ListNamespaces lists available function namespaces
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	var namespaces []string
	c.AddCheckRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	getRequest, err := c.newRequest(http.MethodGet, namespacesPath, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	res, err := c.doRequest(ctx, getRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read namespaces from OpenFaaS on URL: %s", c.GatewayURL.String())
		}
		jsonErr := json.Unmarshal(bytesOut, &namespaces)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse namespaces from OpenFaaS on URL: %s\n%s", c.GatewayURL.String(), jsonErr.Error())
		}
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return namespaces, nil
}
