// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
	types "github.com/openfaas/faas-provider/types"
)

func Test_GetFunctionInfo(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedGetFunctionInfoResponse,
		},
	})

	defer s.Close()
	cliAuth := NewTestAuth(nil)
	proxyClient := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	result, err := proxyClient.GetFunctionInfo(context.Background(), "func-test1", "")
	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
	if expectedGetFunctionInfoResponse != result {
		t.Fatalf("Want: %#v, Got: %#v", expectedGetFunctionInfoResponse, result)
	}
}

func Test_GetFunctionInfo_Not200(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)

	cliAuth := NewTestAuth(nil)
	proxyClient := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	_, err := proxyClient.GetFunctionInfo(context.Background(), "func-test1", "")

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

func Test_GetFunctionInfo_NotFound(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusNotFound)
	cliAuth := NewTestAuth(nil)
	proxyClient := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	functionName := "funct-test"
	_, err := proxyClient.GetFunctionInfo(context.Background(), functionName, "")
	if err == nil {
		t.Fatalf("Error was not returned")
	}

	expectedErrMsg := fmt.Sprintf("No such function: %s", functionName)
	if err.Error() != expectedErrMsg {
		t.Fatalf("Want: %s, Got: %s", expectedErrMsg, err.Error())
	}

}

var expectedGetFunctionInfoResponse = types.FunctionStatus{
	Name:            "func-test1",
	Image:           "image-test1",
	Replicas:        1,
	InvocationCount: 1,
	EnvProcess:      "env-process test1",
}
