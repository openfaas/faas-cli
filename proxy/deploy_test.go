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

type deployProxyTest struct {
	title               string
	mockServerResponses []int
	replace             bool
	update              bool
	expectedOutput      string
	functionName        string
	requests            []test.Request
}

func runDeployProxyTest(t *testing.T, deployTest deployProxyTest) {
	s := test.MockHttpServerStatus(
		t,
		deployTest.mockServerResponses...,
	)
	defer s.Close()

	stdout := test.CaptureStdout(func() {
		DeployFunction(
			"fprocess",
			s.URL,
			"function",
			"image",
			"dXNlcjpwYXNzd29yZA==",
			"language",
			deployTest.replace,
			nil,
			"network",
			[]string{},
			deployTest.update,
			[]string{},
			map[string]string{},
			FunctionResourceRequest{},
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
			mockServerResponses: []int{http.StatusOK, http.StatusOK, http.StatusOK},
			replace:             true,
			update:              false,
			expectedOutput:      `(?m:Deployed)`,
		},
		{
			title:               "404_Deploy",
			mockServerResponses: []int{http.StatusOK, http.StatusOK, http.StatusNotFound},
			replace:             true,
			update:              false,
			expectedOutput:      `(?m:Removing old function)`,
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

	stdout := test.CaptureStdout(func() {
		DeployFunction(
			"fprocess",
			url,
			"function",
			"image",
			"dXNlcjpwYXNzd29yZA==",
			"language",
			false,
			nil,
			"network",
			[]string{},
			false,
			[]string{},
			map[string]string{},
			FunctionResourceRequest{},
		)
	})

	expectedErrMsg := "first path segment in URL cannot contain colon"
	r := regexp.MustCompile(fmt.Sprintf("(?m:%s)", expectedErrMsg))
	if !r.MatchString(stdout) {
		t.Fatalf("Want: %s\nGot: %s", expectedErrMsg, stdout)
	}
}

func runUpdateReplaceProxyTests(t *testing.T, requests []test.Request, replaceUpdateTest deployProxyTest) {
	s := test.MockHttpServer(t, requests)
	defer s.Close()

	stdout := test.CaptureStdout(func() {
		DeployFunction(
			"fprocess",
			s.URL,
			replaceUpdateTest.functionName,
			"image",
			"dXNlcjpwYXNzd29yZA==",
			"lang",
			replaceUpdateTest.replace,
			nil,
			"network",
			[]string{},
			replaceUpdateTest.update,
			[]string{},
			map[string]string{},
			FunctionResourceRequest{},
		)
	})

	r := regexp.MustCompile(fmt.Sprintf("(?m:%s)", replaceUpdateTest.expectedOutput))
	if !r.MatchString(stdout) {
		t.Fatalf("Want: %s\nGot: %s", replaceUpdateTest.expectedOutput, stdout)
	}
}

func Test_DeployFunction_UpdateReplace(t *testing.T) {
	var deployProxyTests = []deployProxyTest{
		{
			title:          "Update_Existing",
			functionName:   "func-test1",
			replace:        false,
			update:         true,
			expectedOutput: "attempting rolling-update", // Success, updating
			requests: []test.Request{
				{
					ResponseStatusCode: http.StatusOK,
					ResponseBody:       expectedListFunctionsResponse,
				},
				{
					ResponseStatusCode: http.StatusOK,
				},
			},
		},
		{
			title:          "Deploy_Existing",
			functionName:   "func-test1",
			replace:        true,
			update:         false,
			expectedOutput: "you must either remove it first, or update it", // Fail, PUTing existing function
			requests: []test.Request{
				{
					ResponseStatusCode: http.StatusOK,
					ResponseBody:       expectedListFunctionsResponse,
				},
			},
		},
		{
			title:          "Update_New",
			functionName:   "new-func",
			replace:        false,
			update:         true,
			expectedOutput: "WARNING! Communication is not secure", // Success, performed POST
			requests: []test.Request{
				{
					ResponseStatusCode: http.StatusOK,
					ResponseBody:       expectedListFunctionsResponse,
				},
				{
					ResponseStatusCode: http.StatusOK,
				},
			},
		},
		{
			title:          "Replace_New",
			functionName:   "new-func",
			replace:        true,
			update:         false,
			expectedOutput: "WARNING! Communication is not secure", // Success, performed POST
			requests: []test.Request{
				{
					ResponseStatusCode: http.StatusOK,
					ResponseBody:       expectedListFunctionsResponse,
				},
				{
					ResponseStatusCode: http.StatusOK,
				},
				{
					ResponseStatusCode: http.StatusOK,
				},
			},
		},
	}
	for _, tst := range deployProxyTests {
		t.Run(tst.title, func(t *testing.T) {
			runUpdateReplaceProxyTests(t, tst.requests, tst)
		})
	}
}
