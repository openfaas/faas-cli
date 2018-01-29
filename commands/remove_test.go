// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
)

func Test_remove(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodDelete,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"remove",
			"--gateway=" + s.URL,
			"test-function",
		})
		faasCmd.Execute()
	})

	expectedStdOut := "Deleting: test-function."
	if found, err := regexp.MatchString(`(?m:`+expectedStdOut+`)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\nGot: %s\nExpected: %s", stdOut, expectedStdOut)
	}
}

func Test_remove_stackYAML(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodDelete,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
		},
	})
	defer s.Close()

	resetForTest()
	funcName := "nodefunc"
	yamlFile = "nodefunc.yml"

	// Cleanup the created directory
	defer func() {
		os.RemoveAll(funcName)
		os.Remove(yamlFile)
	}()

	faasCmd.SetArgs([]string{
		"new",
		"--gateway=" + s.URL,
		"--lang=" + "node",
		funcName,
	})
	faasCmd.Execute()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"remove",
			"--gateway=" + s.URL,
			"--yaml=" + yamlFile,
		})
		faasCmd.Execute()
	})

	expectedStdOut := fmt.Sprintf("Deleting: %s.", funcName)
	if found, err := regexp.MatchString(`(?m:`+expectedStdOut+`)`, stdOut); err != nil || !found {
		t.Fatalf("Tried to match regex '%s' but got: '%s'\n", expectedStdOut, stdOut)
	}
}
