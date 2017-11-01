// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"fmt"
	"net/http"

	"testing"

	"regexp"

	"github.com/openfaas/faas-cli/test"
)

func Test_InvokeFunction(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusOK)
	defer s.Close()

	bytesIn := []byte("test data")
	_, err := InvokeFunction(
		s.URL,
		"function",
		&bytesIn,
		"text/plain",
		[]string{},
	)

	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
}

func Test_InvokeFunction_Not2xx(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusNotFound)
	defer s.Close()

	bytesIn := []byte("test data")
	_, err := InvokeFunction(
		s.URL,
		"function",
		&bytesIn,
		"text/plain",
		[]string{},
	)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:Server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

func Test_InvokeFunction_MissingURLPrefix(t *testing.T) {

	bytesIn := []byte("test data")
	_, err := InvokeFunction(
		"127.0.0.1:8080",
		"function",
		&bytesIn,
		"text/plain",
		[]string{},
	)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	expectedErrMsg := "cannot connect to OpenFaaS on URL:"
	r := regexp.MustCompile(fmt.Sprintf("(?m:%s)", expectedErrMsg))
	if !r.MatchString(err.Error()) {
		t.Fatalf("Want: %s\nGot: %s", expectedErrMsg, err.Error())
	}
}
