// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"context"
	"net/http"
	"regexp"

	"testing"

	"github.com/openfaas/faas-cli/test"
	types "github.com/openfaas/faas-provider/types"
)

func Test_ListFunctions(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedListFunctionsResponse,
		},
	})
	defer s.Close()

	cliAuth := NewTestAuth(nil)
	client := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)
	result, err := client.ListFunctions(context.Background(), "")

	if err != nil {
		t.Fatalf("Error returned: %s", err)
	}
	for k, v := range result {
		if expectedListFunctionsResponse[k] != v {
			t.Fatalf("Expeceted: %#v - Actual: %#v", expectedListFunctionsResponse[k], v)
		}
	}
}

func Test_ListFunctions_Not200(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)

	cliAuth := NewTestAuth(nil)
	client := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)
	_, err := client.ListFunctions(context.Background(), "")

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

var expectedListFunctionsResponse = []types.FunctionStatus{
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
