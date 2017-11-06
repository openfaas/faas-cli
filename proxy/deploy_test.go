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

func Test_DeployFunction(t *testing.T) {
	s := test.MockHttpServerStatus(
		t,
		http.StatusOK, // DeleteFunction
		http.StatusOK, // DeployFunction
	)
	defer s.Close()

	stdout := test.CaptureStdout(func() {
		DeployFunction(
			"fproces",
			s.URL,
			"function",
			"image",
			"language",
			true,
			nil,
			"network",
			[]string{},
			false,
			[]string{},
			map[string]string{},
		)
	})

	r := regexp.MustCompile(`(?m:Deployed.)`)
	if !r.MatchString(stdout) {
		t.Fatalf("Output not matched: %s", stdout)
	}
}

func Test_DeployFunction_Not2xx(t *testing.T) {
	s := test.MockHttpServerStatus(
		t,
		http.StatusOK,       // DeleteFunction
		http.StatusNotFound, // DeployFunction
	)
	defer s.Close()

	stdout := test.CaptureStdout(func() {
		DeployFunction(
			"fproces",
			s.URL,
			"function",
			"image",
			"language",
			true,
			nil,
			"network",
			[]string{},
			false,
			[]string{},
			map[string]string{},
		)
	})

	r := regexp.MustCompile(`(?m:Unexpected status: 404)`)
	if !r.MatchString(stdout) {
		t.Fatalf("Output not matched: %s", stdout)
	}
}

func Test_DeployFunction_MissingURLPrefix(t *testing.T) {
	url := "127.0.0.1:8080"

	stdout := test.CaptureStdout(func() {
		DeployFunction(
			"fprocess",
			url,
			"function",
			"image",
			"language",
			false,
			nil,
			"network",
			[]string{},
			false,
			[]string{},
			map[string]string{},
		)
	})

	expectedErrMsg := "first path segment in URL cannot contain colon"
	r := regexp.MustCompile(fmt.Sprintf("(?m:%s)", expectedErrMsg))
	if !r.MatchString(stdout) {
		t.Fatalf("Want: %s\nGot: %s", expectedErrMsg, stdout)
	}
}
