// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
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

func Test_preRunSecretUpdate_NoArgs_Fails(t *testing.T) {
	res := preRunSecretUpdate(nil, []string{})

	want := "secret name required"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preRunSecretUpdate_MoreThan1Arg_Fails(t *testing.T) {
	res := preRunSecretUpdate(nil, []string{
		"secret1",
		"secret2",
	})

	want := "too many values for secret name"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preRunSecretUpdate_ExtactlyOneArgIsFine(t *testing.T) {
	res := preRunSecretUpdate(nil, []string{
		"secret1",
	})

	if res != nil {
		t.Errorf("expected no validation error, but got %q", res.Error())
	}
}

func Test_SecretUpdateFromStdin(t *testing.T) {
	secretName := "test-secret"
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPut,
			Uri:                "/system/secrets",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       secretName,
		},
	})
	defer s.Close()

	os.Stdin, _ = ioutil.TempFile("", "stdin")
	os.Stdin.WriteString("update-me")
	os.Stdin.Seek(0, 0)
	defer func() {
		os.Remove(os.Stdin.Name())
	}()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"secret",
			"update",
			"--gateway=" + s.URL,
			secretName,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:`+secretName+`)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\nExpected:\n%s\n Got:\n%s", `(?m:`+secretName+`)`, stdOut)
	}
}

func Test_SecretUpdateFromLiteral(t *testing.T) {
	secretName := "test-secret"
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPut,
			Uri:                "/system/secrets",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       secretName,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"secret",
			"update",
			"--gateway=" + s.URL,
			secretName,
			`--from-literal="update-me"`,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:`+secretName+`)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\nExpected:\n%s\n Got:\n%s", `(?m:`+secretName+`)`, stdOut)
	}
}
