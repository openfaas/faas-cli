// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/faas/gateway/requests"
)

func Test_GetFunctionInfo(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedGetFunctionInfoResponse,
		},
	})

	defer s.Close()
	result, err := GetFunctionInfo(s.URL, "func-test1", !tlsNoVerify)
	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
	if expectedGetFunctionInfoResponse != result {
		t.Fatalf("Want: %#v, Got: %#v", expectedGetFunctionInfoResponse, result)
	}
}

func Test_GetFunctionInfo_Not200(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)

	_, err := GetFunctionInfo(s.URL, "func-test1", tlsNoVerify)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

func Test_GetFunctionInfo_MissingURLPrefix(t *testing.T) {
	_, err := GetFunctionInfo("127.0.0.1:8080", "func-test", tlsNoVerify)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	expectedErrMsg := "cannot connect to OpenFaaS on URL:"
	r := regexp.MustCompile(fmt.Sprintf("(?m:%s)", expectedErrMsg))
	if !r.MatchString(err.Error()) {
		t.Fatalf("Want: %s, Got: %s", expectedErrMsg, err.Error())
	}
}

var expectedGetFunctionInfoResponse = requests.Function{
	Name:            "func-test1",
	Image:           "image-test1",
	Replicas:        1,
	InvocationCount: 1,
	EnvProcess:      "env-process test1",
}
