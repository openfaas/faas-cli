// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
)

// ListFunctions list deployed functions
func ListFunctions(gateway string) ([]requests.Function, error) {
	var results []requests.Function

	res, err := http.Get(gateway + "/system/functions")
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case 200:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", gateway)
		}
		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", gateway, jsonErr.Error())
		}

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return results, nil
}
