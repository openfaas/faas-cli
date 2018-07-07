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
		faasCmd.SetArgs([]string{"version"})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:CLI commit: sha-test)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:CLI version: dev)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_addVersion(t *testing.T) {
	version.GitCommit = "sha-test"
	version.Version = "version.tag"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{"version"})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:CLI commit: sha-test)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:CLI version: version.tag)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_addVersion_short_version(t *testing.T) {
	version.Version = "version.tag"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{"version", "--short-version"})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString("^version\\.tag", stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_gateway_uri(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:uri: %s)`, s.URL), stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_gateway_version(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:version: gateway-0.4.3)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_gateway_sha(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:sha: 999a6669148c30adeb64400609953cf59db2fb64)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_gateway_commit(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:commit: Bump faas-swarm to latest)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_name(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:name: faas-swarm)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_orchestration(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:orchestration: swarm)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_version(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:version: provider-0.3.3)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_sha(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_0_8_4_onwards,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:sha: c890cba302d059de8edbef3f3de7fe15444b1ecf)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_gateway_uri_prior_to_0_8_4(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_prior_to_0_8_4,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(fmt.Sprintf(`(?m:uri: %s)`, s.URL), stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_gateway_details_prior_to_0_8_4_should_not_be_displayed(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_prior_to_0_8_4,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:\tversion: $)`, stdOut); err != nil || found {
		t.Fatalf("Output is not as expected for version:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:\tsha: $)`, stdOut); err != nil || found {
		t.Fatalf("Output is not as expected for sha:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:\tcommit: $)`, stdOut); err != nil || found {
		t.Fatalf("Output is not as expected for commit:\n%s", stdOut)
	}
}

func Test_provider_name_prior_to_0_8_4(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_prior_to_0_8_4,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:name: faas-swarm)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_sha_prior_to_0_8_4(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_prior_to_0_8_4,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:sha: c890cba302d059de8edbef3f3de7fe15444b1ecf)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_version_prior_to_0_8_4(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_prior_to_0_8_4,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:version: provider-0.3.3)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_provider_orchestration_prior_to_0_8_4(t *testing.T) {
	resetForTest()
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/info",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       gateway_response_prior_to_0_8_4,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"version",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:orchestration: swarm)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
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
