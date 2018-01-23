// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
	"github.com/spf13/cobra"
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
		deployCmd := newDeployCmd()
		deployCmd.SetArgs([]string{
			"deploy",
			"--gateway=" + s.URL,
			"--image=golang",
			"--name=test-function",
		})
		deployCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:Deployed)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:200 OK)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_ensureUpdateReplaceFlags(t *testing.T) {
	t.Run("With both flags --replace and --update", func(t *testing.T) {
		testCases := [][]string{
			{
				"--update",
				"--replace",
			},
		}
		for _, testCase := range testCases {
			deployCmd := newDeployCmd()
			deployCmd.SetOutput(ioutil.Discard)
			deployCmd.SetArgs(testCase)
			deployCmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
			err := deployCmd.Execute()

			if err != ErrorExclusiveFlagsUpdateReplace {
				t.Errorf("Expected error '%s' does not match actual error '%s'", ErrorExclusiveFlagsUpdateReplace, err)
			}
		}
	})

	t.Run("Deploy with replace", func(t *testing.T) {
		testCases := [][]string{
			{
				"--replace",
			},
			{
				"--replace=true",
			},
		}
		for _, testCase := range testCases {
			deployCmd := newDeployCmd()
			deployCmd.SetOutput(ioutil.Discard)
			deployCmd.SetArgs(testCase)
			deployCmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
			err := deployCmd.Execute()

			if err != nil {
				t.Errorf("error is not nil: %s", err)
			}
			if deployFlags.replace != true {
				t.Errorf("replace is not true as expected")
			}
			if deployFlags.update != false {
				t.Errorf("update is not false as expected")
			}
		}
	})

	t.Run("Deploy with rolling-update", func(t *testing.T) {
		testCases := [][]string{
			{},
			{
				"--update",
			},
			{
				"--update=true",
			},
			{
				"--replace=false",
			},
		}
		for _, testCase := range testCases {
			deployCmd := newDeployCmd()
			deployCmd.SetOutput(ioutil.Discard)
			deployCmd.SetArgs(testCase)
			deployCmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
			err := deployCmd.Execute()

			if err != nil {
				t.Errorf("error is not nil: %s", err)
			}
			if deployFlags.replace != false {
				t.Errorf("replace is not false as expected")
			}
			if deployFlags.update != true {
				t.Errorf("update is not true as expected")
			}
		}
	})
}
