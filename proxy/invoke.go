// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"bytes"
	"os"

	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// InvokeFunction a function
func InvokeFunction(gateway string, name string, bytesIn *[]byte, contentType string, query []string, headers []string, async bool, httpMethod string, tlsInsecure bool, namespace string) (*[]byte, error) {
	var resBytes []byte

	gateway = strings.TrimRight(gateway, "/")

	reader := bytes.NewReader(*bytesIn)

	var disableFunctionTimeout *time.Duration
	client := MakeHTTPClient(disableFunctionTimeout, tlsInsecure)

	qs, qsErr := buildQueryString(query)
	if qsErr != nil {
		return nil, qsErr
	}

	headerMap, headerErr := parseHeaders(headers)
	if headerErr != nil {
		return nil, headerErr
	}

	functionEndpoint := "/function/"
	if async {
		functionEndpoint = "/async-function/"
	}

	httpMethodErr := validateHTTPMethod(httpMethod)
	if httpMethodErr != nil {
		return nil, httpMethodErr
	}

	gatewayURL := gateway + functionEndpoint + name
	if len(namespace) > 0 {
		gatewayURL += "." + namespace
	}
	gatewayURL += qs

	req, err := http.NewRequest(httpMethod, gatewayURL, reader)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	req.Header.Add("Content-Type", contentType)
	// Add additional headers to request
	for name, value := range headerMap {
		req.Header.Add(name, value)
	}

	// Removed by AE - the system-level basic auth secrets should not be transmitted
	// to functions. Functions should implement their own auth.
	// SetAuth(req, gateway)

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
	case http.StatusAccepted:
		fmt.Fprintf(os.Stderr, "Function submitted asynchronously.\n")
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
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
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
				return "", fmt.Errorf("the --query flags must take the form of key=value (= not found)")
			}
			if strings.HasSuffix(queryValue, "=") {
				return "", fmt.Errorf("the --query flag must take the form of: key=value (empty value given, or value ends in =)")
			}
		}
		qs = strings.TrimRight(qs, "&")
	}

	return qs, nil
}

// parseHeaders parses header values from command
func parseHeaders(headers []string) (map[string]string, error) {
	headerMap := make(map[string]string)

	for _, header := range headers {
		headerValues := strings.SplitN(header, "=", 2)
		if len(headerValues) != 2 {
			return headerMap, fmt.Errorf("the --header or -H flag must take the form of key=value")
		}

		name, value := headerValues[0], headerValues[1]
		if name == "" {
			return headerMap, fmt.Errorf("the --header or -H flag must take the form of key=value (empty key given)")
		}

		if value == "" {
			return headerMap, fmt.Errorf("the --header or -H flag must take the form of key=value (empty value given)")
		}

		headerMap[name] = value
	}
	return headerMap, nil
}

// validateMethod validates the HTTP request method
func validateHTTPMethod(httpMethod string) error {
	var allowedMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	}
	helpString := strings.Join(allowedMethods, "/")

	if !contains(allowedMethods, httpMethod) {
		return fmt.Errorf("the --method or -m flag must take one of these values (%s)", helpString)
	}
	return nil
}

func contains(s []string, item string) bool {
	for _, value := range s {
		if value == item {
			return true
		}
	}
	return false
}
