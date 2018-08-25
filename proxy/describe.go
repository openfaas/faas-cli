// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openfaas/faas/gateway/requests"
)

//GetFunctionInfo get an OpenFaaS function information
func GetFunctionInfo(gateway string, functionName string, tlsInsecure bool) (requests.Function, error) {
	var result requests.Function

	gateway = strings.TrimRight(gateway, "/")
	timeout := 60 * time.Second
	client := MakeHTTPClient(&timeout, tlsInsecure)

	getRequest, err := http.NewRequest(http.MethodGet, gateway+"/system/function/"+functionName, nil)
	if err != nil {
		return result, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}
	SetAuth(getRequest, gateway)

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
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return result, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return result, nil
}
