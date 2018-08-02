// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/test"
)

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

func Test_getRegistryAuth_CustomRegistry_NotFound(t *testing.T) {
	wantAuth := ""
	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{},
	}

	result := getRegistryAuth(&configFile1, "my-custom-registry.com/alexellis2/tester")

	if result != wantAuth {
		t.Errorf("want %s (empty), got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_getRegistryAuth_CustomRegistry_Found(t *testing.T) {
	wantAuth := "alexellis2-auth-str"
	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{
			"my-custom-registry.com": authConfig{Auth: wantAuth},
		},
	}

	result := getRegistryAuth(&configFile1, "my-custom-registry.com/alexellis2/tester")

	if result != wantAuth {
		t.Errorf("want %s, got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_getRegistryAuth_NestedGitlabRegistry_Found(t *testing.T) {
	wantAuth := "alexellis2-auth-str"
	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{
			"registry.gitlab.com": authConfig{Auth: wantAuth},
		},
	}

	result := getRegistryAuth(&configFile1, "registry.gitlab.com/alexellis2/tester/function1")

	if result != wantAuth {
		t.Errorf("want %s, got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_getRegistryAuth_CustomRegistry_NoUserPrefix(t *testing.T) {
	hubAuth := "alexellis2-auth-str"
	wantAuth := "alexellis2-registry-local"

	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{
			defaultDockerRegistry: authConfig{Auth: hubAuth},
			"registry.local:5000": authConfig{Auth: wantAuth},
		},
	}

	result := getRegistryAuth(&configFile1, "registry.local:5000/tester")

	if result != wantAuth {
		t.Errorf("registry auth without username - want %s, got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_getRegistryAuth_DockerHub_Found(t *testing.T) {
	wantAuth := "alexellis2-auth-str"
	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{
			defaultDockerRegistry: authConfig{Auth: wantAuth},
		},
	}

	result := getRegistryAuth(&configFile1, "alexellis2/tester")

	if result != wantAuth {
		t.Errorf("want %s, got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_getRegistryAuth_DockerHub_NotFound(t *testing.T) {
	wantAuth := ""
	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{},
	}

	result := getRegistryAuth(&configFile1, "alexellis2/tester")

	if result != "" {
		t.Errorf("want %s (empty), got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_getRegistryAuth_NotRequiredForLocalImage(t *testing.T) {
	wantAuth := ""
	configFile1 := configFile{
		AuthConfigs: map[string]authConfig{
			defaultDockerRegistry: authConfig{Auth: "alexellis2-auth-str"},
		},
	}

	result := getRegistryAuth(&configFile1, "tester:latest")

	if result != "" {
		t.Errorf("want %s (empty), got %s", wantAuth, result)
		t.Fail()
	}
}

func Test_deployFailed(t *testing.T) {

	var failedDeploy = make(map[string]int)
	var containedErrorsCount int
	failedDeploy["example1"] = 100
	failedDeploy["example2"] = 300
	failedDeploy["example3"] = 400
	failedDeploy["example4"] = 500
	err := deployFailed(failedDeploy)
	if err == nil {
		t.Errorf("\nHad to exit with errors!")
		t.Fail()
	}
	for _, theErrorCode := range failedDeploy {
		if strings.Contains(err.Error(), strconv.Itoa(theErrorCode)) {
			containedErrorsCount++
		}
	}
	if containedErrorsCount != len(failedDeploy) {
		t.Errorf("\nWanted: %d number of errors and got: %d!", len(failedDeploy), containedErrorsCount)
		t.Fail()
	}
}
func Test_deploySucceeded(t *testing.T) {
	var succededDeploy = make(map[string]int)
	if err := deployFailed(succededDeploy); err != nil {
		t.Errorf("\nHad to exit with no errors!")
		t.Fail()
	}
}
func Test_badStatusCOde(t *testing.T) {
	okStatusCode := 200
	if badStatusCode(okStatusCode) {
		t.Errorf("\nUnexpected status code - wanted:%d OK!", okStatusCode)
		t.Fail()
	}
	acceptedStatusCode := 202
	if badStatusCode(acceptedStatusCode) {
		t.Errorf("\nUnexpected status code - wanted:%d Accepted!", acceptedStatusCode)
		t.Fail()
	}
	badStatusC := 300
	if !(badStatusCode(badStatusC)) {
		t.Errorf("\nUnexpected status code - wanted: %d but got %d or %d", badStatusC, acceptedStatusCode, okStatusCode)
		t.Fail()
	}
}
