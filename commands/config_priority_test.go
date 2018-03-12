// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"
)

// Test_getGatewayURL tests for priority of URL for gateway over several sources
func Test_getGatewayURL(t *testing.T) {
	defaultValue := "http://127.0.0.1:8080"
	testCases := []struct {
		name       string
		defaultURL string

		yamlURL        string
		argumentURL    string
		environmentURL string
		expectedURL    string
	}{
		{
			name:        "Nothing provided",
			defaultURL:  defaultValue,
			yamlURL:     "",
			argumentURL: "",
			expectedURL: "http://127.0.0.1:8080",
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
		{
			name:           "YAML provided (with defaults) and env-var override",
			defaultURL:     defaultValue,
			yamlURL:        defaultValue,
			environmentURL: "http://remote2:8080",
			argumentURL:    "",
			expectedURL:    "http://remote2:8080",
		},
		{
			name:           "ARG provided with env-var override",
			defaultURL:     defaultValue,
			yamlURL:        defaultValue,
			environmentURL: "http://remote2:8080",
			argumentURL:    "http://remote1:8080",
			expectedURL:    "http://remote1:8080",
		},
	}

	fails := 0
	for _, testCase := range testCases {
		url := getGatewayURL(testCase.argumentURL, testCase.defaultURL, testCase.yamlURL, testCase.environmentURL)
		if url != testCase.expectedURL {
			t.Logf("gatewayURL %s\nwant: %s, got: %s", testCase.name, testCase.expectedURL, url)
			fails++
		}
	}
	if fails > 0 {
		t.Fail()
	}
}
