// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/alexellis/hmac"
	"github.com/openfaas/faas-cli/test"
)

func Test_invoke(t *testing.T) {
	expectedInvokeResponse := "response-test-data"
	funcName := "test-1"

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPost,
			Uri:                "/function/" + funcName,
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedInvokeResponse,
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

	if found, err := regexp.MatchString(`(?m:`+expectedInvokeResponse+`)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\nExpected:\n%s\n Got:\n%s", `(?m:`+expectedInvokeResponse+`)`, stdOut)
	}

}

func Test_async_invoke(t *testing.T) {
	funcName := "test-1"

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPost,
			Uri:                "/async-function/" + funcName,
			ResponseStatusCode: http.StatusAccepted,
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
			"--async",
			funcName,
		})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:)`, stdOut); err != nil || !found {
		t.Fatalf("Async output is not as expected:\nExpected:\n%s\n Got:\n%s", `(?m:)`, stdOut)
	}

}

func Test_generateSignedHeader(t *testing.T) {

	var generateTestcases = []struct {
		title       string
		message     []byte
		key         string
		headerName  string
		expectedSig string
		expectedErr bool
	}{
		{
			title:       "Header with empty key",
			message:     []byte("This is a message"),
			key:         "",
			headerName:  "HeaderSet",
			expectedSig: "HeaderSet=sha1=cdefd604e685e5c8b31fbcf6621a6e8282770dfe",
			expectedErr: false,
		},
		{
			title:       "Key with empty Header",
			message:     []byte("This is a message"),
			key:         "KeySet",
			headerName:  "",
			expectedSig: "",
			expectedErr: true,
		},
		{
			title:       "Header & key with empty message",
			message:     []byte(""),
			key:         "KeySet",
			headerName:  "HeaderSet",
			expectedSig: "HeaderSet=sha1=33dcd94ffaf13fce58615585c030c1a39d100b3c",
			expectedErr: false,
		},
		{
			title:       "Header with empty message & key",
			message:     []byte(""),
			key:         "",
			headerName:  "HeaderSet",
			expectedSig: "HeaderSet=sha1=fbdb1d1b18aa6c08324b7d64b71fb76370690e1d",
			expectedErr: false,
		},
	}
	for _, test := range generateTestcases {
		t.Run(test.title, func(t *testing.T) {
			sig, err := generateSignedHeader(test.message, test.key, test.headerName)

			if sig != test.expectedSig {
				t.Fatalf("error generating signature, wanted: %s, got %s", test.expectedSig, sig)
			}
			if (err != nil) != test.expectedErr {
				t.Fatalf("error generating expected error: %v, got: %v", err != nil, test.expectedErr)
			}

			if test.expectedErr == false {

				encodedHash := strings.SplitN(test.expectedSig, "=", 2)

				invalid := hmac.Validate(test.message, encodedHash[1], test.key)

				if invalid != nil {
					t.Fatalf("expected no error, but got: %s", invalid.Error())
				}
			}
		})
	}
}

func Test_missingSignFlag(t *testing.T) {

	var signtestcases = []struct {
		title       string
		hdr         string
		key         string
		expectedRes bool
	}{
		{
			title:       "Header and key",
			hdr:         "Header",
			key:         "Key",
			expectedRes: false,
		},
		{
			title:       "Header without key",
			hdr:         "Header",
			key:         "",
			expectedRes: true,
		},
		{
			title:       "Key without Header",
			hdr:         "",
			key:         "Key",
			expectedRes: true,
		},
		{
			title:       "No Key No Header",
			hdr:         "",
			key:         "",
			expectedRes: false,
		},
	}
	for _, test := range signtestcases {
		t.Run(test.title, func(t *testing.T) {

			res := missingSignFlag(test.hdr, test.key)

			if res != test.expectedRes {
				t.Fatalf("error testing ability to sign, wanted: %v, got %v", test.expectedRes, res)
			}
		})
	}
}
