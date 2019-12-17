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
		[]string{},
		false,
		http.MethodPost,
		tlsNoVerify,
		"",
	)

	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
}

func Test_InvokeFunction_Async(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusAccepted)
	defer s.Close()

	bytesIn := []byte("test data")
	_, err := InvokeFunction(
		s.URL,
		"function",
		&bytesIn,
		"text/plain",
		[]string{},
		[]string{},
		true,
		http.MethodPost,
		tlsNoVerify,
		"",
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
		[]string{},
		false,
		http.MethodPost,
		tlsNoVerify,
		"",
	)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code)`)
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
		[]string{},
		false,
		http.MethodPost,
		tlsNoVerify,
		"",
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

func Test_ParseHeaders(t *testing.T) {
	testcases := []struct {
		Name   string
		Input  []string
		Output map[string]string
	}{
		{
			Name:  "Header with key-value pair as value",
			Input: []string{`X-Hub-Signature="sha1="shashashaebaf43""`, "X-Hub-Signature-1=sha1=shashashaebaf43"},
			Output: map[string]string{"X-Hub-Signature": `"sha1="shashashaebaf43""`,
				"X-Hub-Signature-1": "sha1=shashashaebaf43"},
		},
		{
			Name:   "Header with normal values",
			Input:  []string{`X-Hub-Signature="shashashaebaf43"`, "X-Hub-Signature-1=shashashaebaf43"},
			Output: map[string]string{"X-Hub-Signature": `"shashashaebaf43"`, "X-Hub-Signature-1": "shashashaebaf43"},
		},
		{
			Name:   "Header with base64 string value",
			Input:  []string{`X-Hub-Signature="shashashaebaf43="`},
			Output: map[string]string{"X-Hub-Signature": `"shashashaebaf43="`},
		},
	}

	for _, testcase := range testcases {
		output, err := parseHeaders(testcase.Input)

		if err != nil {
			t.Fatalf("Testcase %s failed : %s", testcase.Name, err.Error())
		}

		if err == nil && !compareMaps(testcase.Output, output) {
			t.Fatalf("Testcase %s failed. Want: %s, Got: %s", testcase.Name, testcase.Output, output)
		}
	}
}

func compareMaps(mapA map[string]string, mapB map[string]string) bool {
	for key, valueA := range mapA {
		valueB, exists := mapB[key]

		if !exists {
			return false
		}

		if exists && valueA != valueB {
			return false
		}
	}
	return true
}
