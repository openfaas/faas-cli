// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
	types "github.com/openfaas/faas-provider/types"
)

func Test_list(t *testing.T) {
	expectedListResponse := []types.FunctionStatus{
		{
			Name:            "function-test-1",
			Image:           "image-test-1",
			Replicas:        1,
			InvocationCount: 3,
		},
		{
			Name:            "function-test-2",
			Image:           "image-test-2",
			Replicas:        3,
			InvocationCount: 999999,
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedListResponse,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"list",
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	matches := regexp.MustCompile(`(?m:function-test-[12])`).FindAllStringSubmatch(stdOut, 2)
	if len(matches) != 2 {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_list_errors(t *testing.T) {

	resetForTest()

	faasCmd.SetArgs([]string{
		"list", "--gateway", "bad-gateway",
	})
	err := faasCmd.Execute()

	if err == nil {
		t.Fatal("No error found while testing bad gateway")
	}

	resetForTest()

	faasCmd.SetArgs([]string{
		"list", "--yaml", "non-existant-yaml",
	})
	err = faasCmd.Execute()

	if err == nil {
		t.Fatal("No error found while testing missing yaml")
	}
}
