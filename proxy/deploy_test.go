// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"fmt"
	"net/http"

	"testing"

	"regexp"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/faas/gateway/requests"
)

const tlsNoVerify = true

type deployProxyTest struct {
	title               string
	mockServerResponses []int
	replace             bool
	update              bool
	expectedOutput      string
}

func runDeployProxyTest(t *testing.T, deployTest deployProxyTest) {
	s := test.MockHttpServerStatus(
		t,
		deployTest.mockServerResponses...,
	)
	defer s.Close()

	req := requests.CreateFunctionRequest{
		EnvProcess:             "fprocess",
		Image:                  "image",
		RegistryAuth:           "dXNlcjpwYXNzd29yZA==",
		Network:                "network",
		Service:                "function",
		EnvVars:                nil,
		Constraints:            []string{},
		Secrets:                []string{},
		Labels:                 &map[string]string{},
		Annotations:            &map[string]string{},
		ReadOnlyRootFilesystem: false,
	}
	stdout := test.CaptureStdout(func() {
		DeployFunction(
			req,
			s.URL,
			"language",
			deployTest.replace,
			deployTest.update,
			FunctionResourceRequest{},
			tlsNoVerify,
		)
	})

	r := regexp.MustCompile(deployTest.expectedOutput)
	if !r.MatchString(stdout) {
		t.Fatalf("Output not matched: %s", stdout)
	}
}

func Test_RunDeployProxyTests(t *testing.T) {
	var deployProxyTests = []deployProxyTest{
		{
			title:               "200_Deploy",
			mockServerResponses: []int{http.StatusOK, http.StatusOK},
			replace:             true,
			update:              false,
			expectedOutput:      `(?m:Deployed)`,
		},
		{
			title:               "404_Deploy",
			mockServerResponses: []int{http.StatusOK, http.StatusNotFound},
			replace:             true,
			update:              false,
			expectedOutput:      `(?m:Unexpected status: 404)`,
		},
		{
			title:               "UpdateFailedDeployed",
			mockServerResponses: []int{http.StatusNotFound, http.StatusOK},
			replace:             false,
			update:              true,
			expectedOutput:      `(?m:Deployed)`,
		},
	}
	for _, tst := range deployProxyTests {
		t.Run(tst.title, func(t *testing.T) {
			runDeployProxyTest(t, tst)
		})
	}
}

func Test_DeployFunction_MissingURLPrefix(t *testing.T) {
	url := "127.0.0.1:8080"

	req := requests.CreateFunctionRequest{
		EnvProcess:             "fprocess",
		Image:                  "image",
		RegistryAuth:           "dXNlcjpwYXNzd29yZA==",
		Network:                "network",
		Service:                "function",
		EnvVars:                nil,
		Constraints:            []string{},
		Secrets:                []string{},
		Labels:                 &map[string]string{},
		Annotations:            &map[string]string{},
		ReadOnlyRootFilesystem: false,
	}
	stdout := test.CaptureStdout(func() {
		DeployFunction(
			req,
			url,
			"language",
			false,
			false,
			FunctionResourceRequest{},
			tlsNoVerify,
		)
	})

	expectedErrMsg := "first path segment in URL cannot contain colon"
	r := regexp.MustCompile(fmt.Sprintf("(?m:%s)", expectedErrMsg))
	if !r.MatchString(stdout) {
		t.Fatalf("Want: %s\nGot: %s", expectedErrMsg, stdout)
	}
}
