// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
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

//GetFunctionInfo get an OpenFaaS function information
func (c *Client) GetFunctionInfo(ctx context.Context, functionName string, namespace string) (types.FunctionStatus, error) {
	var (
		result types.FunctionStatus
		err    error
	)

	functionPath := fmt.Sprintf("%s/%s", functionPath, functionName)
	if len(namespace) > 0 {
		functionPath, err = addQueryParams(functionPath, map[string]string{namespaceKey: namespace})
		if err != nil {
			return result, err
		}
	}

	getRequest, err := c.newRequest(http.MethodGet, functionPath, nil)
	if err != nil {
		return result, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	res, err := c.doRequest(ctx, getRequest)
	if err != nil {
		return result, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return result, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", c.GatewayURL.String())
		}

		jsonErr := json.Unmarshal(bytesOut, &result)
		if jsonErr != nil {
			return result, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", c.GatewayURL.String(), jsonErr.Error())
		}
	case http.StatusUnauthorized:
		return result, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	case http.StatusNotFound:
		return result, fmt.Errorf("No such function: %s", functionName)
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return result, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return result, nil
}
