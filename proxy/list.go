// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	types "github.com/openfaas/faas-provider/types"
)

// ListFunctions list deployed functions
func ListFunctions(gateway string, tlsInsecure bool, namespace string) ([]types.FunctionStatus, error) {
	return ListFunctionsToken(gateway, tlsInsecure, "", namespace)
}

// ListFunctionsToken list deployed functions with a token as auth
func ListFunctionsToken(gateway string, tlsInsecure bool, token string, namespace string) ([]types.FunctionStatus, error) {
	var results []types.FunctionStatus

	gateway = strings.TrimRight(gateway, "/")
	client := MakeHTTPClient(&defaultCommandTimeout, tlsInsecure)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	getEndpoint, err := createSystemEndpoint(gateway, namespace)
	if err != nil {
		return results, err
	}

	getRequest, err := http.NewRequest(http.MethodGet, getEndpoint, nil)

	if len(token) > 0 {
		SetToken(getRequest, token)
	} else {
		SetAuth(getRequest, gateway)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	res, err := client.Do(getRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", gateway)
		}
		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", gateway, jsonErr.Error())
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
