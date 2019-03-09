// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"io/ioutil"

	"github.com/openfaas/faas-cli/test"
)

func Test_preRunSecretCreate_NoArgs_Fails(t *testing.T) {
	res := preRunSecretCreate(nil, []string{})

	want := "secret name required"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preRunSecretCreate_MoreThan1Arg_Fails(t *testing.T) {
	res := preRunSecretCreate(nil, []string{
		"secret1",
		"secret2",
	})

	want := "too many values for secret name"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preRunSecretCreate_ExtactlyOneArgIsFine(t *testing.T) {
	res := preRunSecretCreate(nil, []string{
		"secret1",
	})

	if res != nil {
		t.Errorf("expected no validation error, but got %q", res.Error())
	}
}

func Test_SecretCreateFromStdin(t *testing.T) {
	secretName := "test-secret"
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPost,
			Uri:                "/system/secrets",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       secretName,
		},
	})
	defer s.Close()

	os.Stdin, _ = ioutil.TempFile("", "stdin")
	os.Stdin.WriteString("hello")
	os.Stdin.Seek(0, 0)
	defer func() {
		os.Remove(os.Stdin.Name())
	}()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"secret",
			"create",
			"--gateway=" + s.URL,
			secretName,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:`+secretName+`)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\nExpected:\n%s\n Got:\n%s", `(?m:`+secretName+`)`, stdOut)
	}
}

func Test_SecretCreateFromLiteral(t *testing.T) {
	secretName := "test-secret"
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPost,
			Uri:                "/system/secrets",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       secretName,
		},
	})
	defer s.Close()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"secret",
			"create",
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

func Test_validateSecretName(t *testing.T) {

	testcases := []struct {
		Name       string
		SecretName string
		Err        error
	}{
		{
			Name:       "Valid secret with hyphen",
			SecretName: "api-key-secret",
			Err:        nil,
		},
		{
			Name:       "Valid secret without hyphen",
			SecretName: "apikeysecret",
			Err:        nil,
		},
		{
			Name:       "Valid secret with hyphen and number",
			SecretName: "api-key-secret-123",
			Err:        nil,
		},
		{
			Name:       "Invalid secret name",
			SecretName: "api_key_@secret",
			Err:        fmt.Errorf(invalidSecretNameMessage, "api_key_@secret"),
		},
		{
			Name:       "Invalid secret name with number",
			SecretName: "12api_key_secret",
			Err:        fmt.Errorf(invalidSecretNameMessage, "12api_key_secret"),
		},
	}

	for _, test := range testcases {
		isValid, err := validateSecretName(test.SecretName)
		if isValid {
			if err != nil {
				t.Errorf("testcase %s failed, returned error for valid secret %s", test.Name, test.SecretName)
			}
		} else {
			if err.Error() != test.Err.Error() {
				t.Errorf("testcase %s failed, expected: %s, got: %s", test.Name, test.Err.Error(), err.Error())
			}
		}
	}
}
