// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"context"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"

	types "github.com/openfaas/faas-provider/types"
)

// ListFunctions list deployed functions
func (c *Client) ListFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {
	var (
		results      []types.FunctionStatus
		listEndpoint string
		err          error
	)

	c.AddCheckRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	listEndpoint = systemPath
	if len(namespace) > 0 {
		listEndpoint, err = addQueryParams(listEndpoint, map[string]string{namespaceKey: namespace})
		if err != nil {
			return results, err
		}
	}

	getRequest, err := c.newRequest(http.MethodGet, listEndpoint, nil)
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
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", c.GatewayURL.String())
		}
		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", c.GatewayURL.String(), jsonErr.Error())
		}
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return results, nil
}
