// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"net/http"

	"testing"

	"regexp"

	"github.com/openfaas/faas-cli/test"
)

func Test_InvokeFunction(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusOK)
	defer s.Close()

	bytesIn := []byte("test data")
	query := []string{"ascii=<key: 0x90>", "key=foo & bar"}
	_, err := InvokeFunction(
		s.URL,
		"function",
		&query,
		&bytesIn,
		"text/plain",
	)

	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
}

func Test_InvokeFunction_Not2xx(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusNotFound)
	defer s.Close()

	bytesIn := []byte("test data")
	query := []string{"key=val"}
	_, err := InvokeFunction(
		s.URL,
		"function",
		&query,
		&bytesIn,
		"text/plain",
	)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:Server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

func Test_buildURL(t *testing.T) {

	baseURL := "http://localhost:8080/function/function"
	query := []string{"ascii=<key: 0x90>", "key=foo & bar"}
	fullURL := baseURL + "?ascii=%3Ckey%3A+0x90%3E&key=foo+%26+bar"
	u, err := buildURL(baseURL, &query)
	if err != nil {
		if u != fullURL {
			t.Fatalf("building the URL failed")
		}
	}

	query = []string{"no_equal_sign"}
	u, err = buildURL(baseURL, &query)
	if err == nil {
		t.Fatalf("Error was not returned")
	}

	baseURL = "http://127.0.0.%31:8080"
	u, err = buildURL(baseURL, &query)
	if err == nil {
		t.Fatalf("Error was not returned")
	}
}
