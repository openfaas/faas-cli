// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
)

func Test_getGatewayURL(t *testing.T) {
	defaultValue := "http://localhost:8080"
	testCases := []struct {
		name        string
		defaultURL  string
		yamlURL     string
		argumentURL string
		expectedURL string
	}{
		{
			name:        "Nothing provided",
			defaultURL:  defaultValue,
			yamlURL:     "",
			argumentURL: "",
			expectedURL: "http://localhost:8080",
		},
		{
			name:        "Only YAML provided",
			defaultURL:  defaultValue,
			yamlURL:     "http://remote1:8080",
			argumentURL: "",
			expectedURL: "http://remote1:8080",
		},
		{
			name:        "Only argument override",
			defaultURL:  defaultValue,
			yamlURL:     "",
			argumentURL: "http://remote2:8080",
			expectedURL: "http://remote2:8080",
		},
		{
			name:        "Prioritize argument over YAML when argument is not default",
			defaultURL:  defaultValue,
			yamlURL:     "http://remote-yml:8080",
			argumentURL: "http://remote-arg:8080",
			expectedURL: "http://remote-arg:8080",
		},
		{
			name:        "When argument is default use YAML",
			defaultURL:  defaultValue,
			yamlURL:     "http://remote-yml:8080",
			argumentURL: defaultValue,
			expectedURL: "http://remote-yml:8080",
		},
	}

	fails := 0
	for _, testCase := range testCases {
		url := getGatewayURL(testCase.argumentURL, testCase.defaultURL, testCase.yamlURL)
		if url != testCase.expectedURL {
			t.Logf("gatewayURL %s\nwant: %s, got: %s", testCase.name, testCase.expectedURL, url)
			fails++
		}
	}
	if fails > 0 {
		t.Fail()
	}
}

func Test_deploy(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPut,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"deploy",
			"--gateway=" + s.URL,
			"--image=golang",
			"--name=test-function",
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:Deployed)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:200 OK)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}
