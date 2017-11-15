// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"bytes"

	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// InvokeFunction a function
func InvokeFunction(gateway string, name string, bytesIn *[]byte, contentType string, query []string) (*[]byte, error) {
	var resBytes []byte

	gateway = strings.TrimRight(gateway, "/")

	reader := bytes.NewReader(*bytesIn)

	var timeout *time.Duration
	client := MakeHTTPClient(timeout)

	qs, qsErr := buildQueryString(query)
	if qsErr != nil {
		return nil, qsErr
	}

	gatewayURL := gateway + "/function/" + name + qs

	req, err := http.NewRequest(http.MethodPost, gatewayURL, reader)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	req.Header.Add("Content-Type", contentType)
	SetAuth(req, gateway)

	res, err := client.Do(req)

	if err != nil {
		fmt.Println()
		fmt.Println(err)
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		var readErr error
		resBytes, readErr = ioutil.ReadAll(res.Body)
		if readErr != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s %s", gateway, readErr)
		}
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return &resBytes, nil
}

func buildQueryString(query []string) (string, error) {
	qs := ""

	if len(query) > 0 {
		qs = "?"
		for _, queryValue := range query {
			qs = qs + queryValue + "&"
			if strings.Contains(queryValue, "=") == false {
				return "", fmt.Errorf("The --query flags must take the form of key=value (= not found)")
			}
			if strings.HasSuffix(queryValue, "=") {
				return "", fmt.Errorf("The --query flag must take the form of: key=value (empty value given, or value ends in =)")
			}
		}
		qs = strings.TrimRight(qs, "&")
	}

	return qs, nil
}
