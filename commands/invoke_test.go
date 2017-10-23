// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"net/http"
	"os"
	"regexp"
	"testing"

	"io/ioutil"

	"github.com/openfaas/faas-cli/test"
)

func Test_invoke(t *testing.T) {
	expected_invoke_response := "response-test-data"
	funcName := "test-1"

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPost,
			Uri:                "/function/" + funcName,
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expected_invoke_response,
		},
	})
	defer s.Close()

	os.Stdin, _ = ioutil.TempFile("", "stdin")
	os.Stdin.WriteString("test-data")
	os.Stdin.Seek(0, 0)
	defer func() {
		os.Remove(os.Stdin.Name())
	}()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"invoke",
			"--gateway=" + s.URL,
			funcName,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:`+expected_invoke_response+`)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}
