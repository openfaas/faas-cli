// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/json"
	"io/ioutil"
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

func parseReqBuffer(buf []byte) (service string, env map[string]interface{}, err error) {
	var body map[string]interface{}

	err = json.Unmarshal(buf, &body)
	if err != nil {
		return
	}
	service = body["service"].(string)
	env = body["envVars"].(map[string]interface{})
	return
}

func Test_deploy_envFiles(t *testing.T) {
	expected := map[string]map[string]string{
		"func1": {
			"db_host": "192.168.10.2",
			"db_name": "func1_db",
			"db_user": "func1_usr",
		},
		"func2": {
			"db_host":    "localhost",
			"db_name":    "func2_db",
			"db_user":    "func2_user",
			"redis_host": "192.168.0.1",
			"redis_pass": "123456",
		},
	}

	var buffers [][]byte
	var reqs []test.Request

	for i := 0; i < len(expected); i++ {
		reqs = append(reqs, test.Request{
			Method:             http.MethodPut,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			Hook: func(r *http.Request) {
				buf, err := ioutil.ReadAll(r.Body)
				if err != nil {
					return
				}
				buffers = append(buffers, buf)
			},
		})
	}

	s := test.MockHttpServer(t, reqs)
	defer s.Close()

	faasCmd.SetArgs([]string{
		"deploy",
		"--gateway=" + s.URL,
		"-f",
		"./testdata/deploy/stack.yml",
		"--environment-file",
		"./testdata/deploy/test-env-files.yml",
	})

	err := faasCmd.Execute()

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(expected); i++ {
		service, env, err := parseReqBuffer(buffers[i])
		if err != nil {
			t.Fatal(err)
		}

		exp := expected[service]
		
		if len(exp) != len(env) {
			t.Fatal("Invalid environment variable values")
		}

		for k, v := range exp {
			if env[k] != v {
				t.Fatal("Invalid environment variable values")
			}
		}
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
