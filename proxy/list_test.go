// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"net/http"
	"regexp"

	"testing"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/faas/gateway/requests"
)

func Test_ListFunctions(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedListFunctionsResponse,
		},
	})
	defer s.Close()

	result, err := ListFunctions(s.URL)

	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
	for k, v := range result {
		if expectedListFunctionsResponse[k] != v {
			t.Fatal("Expeceted: %#v - Actual: %#v", expectedListFunctionsResponse[k], v)
		}
	}
}

func Test_ListFunctions_Not200(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)

	_, err := ListFunctions(s.URL)

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:Server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

var expectedListFunctionsResponse = []requests.Function{
	{
		Name:            "func-test1",
		Image:           "image-test1",
		Replicas:        1,
		InvocationCount: 1,
		EnvProcess:      "env-process test1",
	},
	{
		Name:            "func-test2",
		Image:           "image-test2",
		Replicas:        2,
		InvocationCount: 2,
		EnvProcess:      "env-process test2",
	},
}
