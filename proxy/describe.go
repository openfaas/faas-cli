// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/openfaas/faas/gateway/requests"
)

//GetFunctionInfo get an OpenFaaS function information
func GetFunctionInfo(gateway string, functionName string, tlsInsecure bool) (requests.Function, error) {
	return GetFunctionInfoToken(gateway, functionName, tlsInsecure, "")
}

//GetFunctionInfoToken get a function information with a token as auth
func GetFunctionInfoToken(gateway string, functionName string, tlsInsecure bool, token string) (requests.Function, error) {
	var result requests.Function

	client := MakeHTTPClient(&defaultCommandTimeout, tlsInsecure)

	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		return result, fmt.Errorf("invalid gateway URL: %s", gateway)
	}
	gatewayURL.Path = path.Join(gatewayURL.Path, "/system/function/", functionName)

	getRequest, err := http.NewRequest(http.MethodGet, gatewayURL.String(), nil)
	if err != nil {
		return result, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	if len(token) > 0 {
		SetToken(getRequest, token)
	} else {
		SetAuth(getRequest, gateway)
	}

	res, err := client.Do(getRequest)
	if err != nil {
		return result, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return result, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", gateway)
		}

		jsonErr := json.Unmarshal(bytesOut, &result)
		if jsonErr != nil {
			return result, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", gateway, jsonErr.Error())
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
