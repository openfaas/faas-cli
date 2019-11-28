// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"regexp"
	"testing"

	"fmt"
	"net/http"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/faas-cli/version"
)

func Test_addVersionDev(t *testing.T) {
	version.GitCommit = "sha-test"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--warn-update=false",
		})
		faasCmd.Execute()
	})

	expected := "commit:  sha-test"
	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:%s)`, expected), stdOut); err != nil || !found {
		t.Fatalf("Commit is not as expected - want: %s, got:\n%s", expected, stdOut)
	}

	expected = "version: dev"
	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:%s)`, expected), stdOut); err != nil || !found {
		t.Fatalf("Version is not as expected - want: %s, got: %s", expected, stdOut)
	}
}

func Test_addVersion(t *testing.T) {
	version.GitCommit = "sha-test"
	version.Version = "version.tag"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--warn-update=false",
		})
		faasCmd.Execute()
	})

	expected := "commit:  sha-test"
	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:%s)`, expected), stdOut); err != nil || !found {
		t.Fatalf("Commit is not as expected:\n%s", stdOut)
	}

	expected = "version: version.tag"
	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:%s)`, expected), stdOut); err != nil || !found {
		t.Fatalf("Version is not as expected - want: %s, got: %s", expected, stdOut)
	}
}

func Test_addVersion_short_version(t *testing.T) {
	version.Version = "version.tag"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--warn-update=false",
			"--short-version",
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString("^version\\.tag", stdOut); err != nil || !found {
		t.Fatalf("Version is not as expected - want: %s, got: %s", version.Version, stdOut)
	}
}

func Test_gateway_and_provider_information(t *testing.T) {
	var testCases = []struct {
		responseBody string
		params       []struct {
			name  string
			value string
		}
	}{
		{
			responseBody: gateway_response_0_8_4_onwards,
			params: []struct {
				name  string
				value string
			}{
				{"gateway version", "version: gateway-0.4.3"},
				{"gateway sha", "sha:     999a6669148c30adeb64400609953cf59db2fb64"},
				{"gateway commit", "commit:  Bump faas-swarm to latest"},
				{"provider name", "name:          faas-swarm"},
				{"provider orchestration", "orchestration: swarm"},
				{"provider version", "version:       provider-0.3.3"},
				{"provider sha", "sha:           c890cba302d059de8edbef3f3de7fe15444b1ecf"},
			},
		},
		{
			responseBody: gateway_response_prior_to_0_8_4,
			params: []struct {
				name  string
				value string
			}{
				{"provider name", "name:          faas-swarm"},
				{"provider orchestration", "orchestration: swarm"},
				{"provider version", "version:       provider-0.3.3"},
				{"provider sha", "sha:           c890cba302d059de8edbef3f3de7fe15444b1ecf"},
			},
		},
	}

	for _, testCase := range testCases {
		stdOut, _ := executeVersionCmd(t, testCase.responseBody)

		for _, param := range testCase.params {
			t.Run(param.name, func(t *testing.T) {
				if found, err := regexp.MatchString(fmt.Sprintf(`(?m:%s)`, param.value), stdOut); err != nil || !found {
					t.Fatalf("%s is not as expected - want: `%s` got:\n`%s`", param.name, param.value, stdOut)
				}
			})
		}
	}
}

func Test_gateway_uri(t *testing.T) {
	stdOut, gatewayUri := executeVersionCmd(t, gateway_response_0_8_4_onwards)

	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:uri:     %s)`, gatewayUri), stdOut); err != nil || !found {
		t.Fatalf("Gateway uri is not as expected - want: %s, got:\n%s", gatewayUri, stdOut)
	}
}

func Test_gateway_uri_prior_to_0_8_4(t *testing.T) {
	stdOut, gatewayUri := executeVersionCmd(t, gateway_response_prior_to_0_8_4)

	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:uri:     %s)`, gatewayUri), stdOut); err != nil || !found {
		t.Fatalf("Gateway uri is not as expected - want: %s, got:\n%s", gatewayUri, stdOut)
	}
}

func Test_gateway_details_prior_to_0_8_4_should_not_be_displayed(t *testing.T) {
	stdOut, _ := executeVersionCmd(t, gateway_response_prior_to_0_8_4)

	if found, err := regexp.MatchString(`(?m:\tversion: $)`, stdOut); err != nil || found {
		t.Fatalf("Version is not as expected - want is not to be there, got:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:\tsha: $)`, stdOut); err != nil || found {
		t.Fatalf("Sha is not as expected - want it not to be there, got:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:\tcommit: $)`, stdOut); err != nil || found {
		t.Fatalf("Commit is not as expected for commit - want it not to be there, got:\n%s", stdOut)
	}
}

func executeVersionCmd(t *testing.T, responseBody string) (versionInfo string, gatewayUri string) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       responseBody,
		},
	})

	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
			"--warn-update=false",
		})
		faasCmd.Execute()
	})

	return stdOut, s.URL
}

const gateway_response_0_8_4_onwards = `{
  "provider": {
    "provider": "faas-swarm",
    "orchestration": "swarm",
    "version": {
      "sha": "c890cba302d059de8edbef3f3de7fe15444b1ecf",
      "release": "provider-0.3.3"
    }
  },
  "version": {
    "sha": "999a6669148c30adeb64400609953cf59db2fb64",
    "release": "gateway-0.4.3",
    "commit_message": "Bump faas-swarm to latest"
  } 
}`

const gateway_response_prior_to_0_8_4 = `{
  "provider": "faas-swarm",
  "version": {
    "sha": "c890cba302d059de8edbef3f3de7fe15444b1ecf",
    "release": "provider-0.3.3"
  },
  "orchestration": "swarm"
}`
