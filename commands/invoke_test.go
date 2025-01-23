// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"testing"

	"io/ioutil"

	"github.com/openfaas/faas-cli/test"
)

func Test_invoke(t *testing.T) {
	expectedInvokeResponse := "response-test-data"
	funcName := "test-1"

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodPost,
			Uri:                "/function/" + funcName + ".openfaas-fn",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedInvokeResponse,
		},
	})
	defer s.Close()

	os.Stdin, _ = os.CreateTemp("", "stdin")
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
			Uri:                "/async-function/" + funcName + ".openfaas-fn",
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

func Test_generateHeader(t *testing.T) {
	var tests = []struct {
		name    string
		message []byte
		key     string
		want    string
	}{
		{
			name:    "Empty key",
			message: []byte("This is a message"),
			key:     "",
			want:    "sha256=7fb67a61acd7a9fa2541bbde51cef1bd4086a5a3acec0a0c821b40e06e824cfc",
		},
		{
			name:    "Key with empty message",
			message: []byte(""),
			key:     "KeySet",
			want:    "sha256=51846d8900847a40a129743c98742c83a56c3cbc4f5aec188d7eb2de629d11df",
		},
		{
			name:    "Empty key and message",
			message: []byte(""),
			key:     "",
			want:    "sha256=b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := generateSignature(test.message, test.key)

			if got != test.want {
				t.Fatalf("error generating signature, want: %s, got %s", test.want, got)
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

func Test_parseHeaders_valid(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  http.Header
	}{
		{
			name:  "Header with key-value pair as value",
			input: []string{`X-Hub-Signature="sha1="shashashaebaf43""`, "X-Hub-Signature-1=sha1=shashashaebaf43"},
			want: http.Header{
				"X-Hub-Signature":   []string{`"sha1="shashashaebaf43""`},
				"X-Hub-Signature-1": []string{"sha1=shashashaebaf43"},
			},
		},
		{
			name:  "Header with normal values",
			input: []string{`X-Hub-Signature="shashashaebaf43"`, "X-Hub-Signature-1=shashashaebaf43"},
			want:  http.Header{"X-Hub-Signature": []string{`"shashashaebaf43"`}, "X-Hub-Signature-1": []string{"shashashaebaf43"}},
		},
		{
			name:  "Header with base64 string value",
			input: []string{`X-Hub-Signature="shashashaebaf43="`},
			want:  http.Header{"X-Hub-Signature": []string{`"shashashaebaf43="`}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseHeaders(test.input)

			if err != nil {
				t.Fatalf("want: %s, got error: %s", test.want, err)
			}

			if err == nil && !reflect.DeepEqual(test.want, got) {
				t.Fatalf("want: %s, got: %s", test.want, got)
			}
		})
	}
}

func Test_parseHeaders_invalid(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{
			name:  "Invalid header string",
			input: []string{"invalid_header"},
		},
		{
			name:  "Empty header key",
			input: []string{"=shashashaebaf43"},
		},
		{
			name:  "Empty header value",
			input: []string{"X-Hub-Signature="},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseHeaders(test.input)

			if err == nil {
				t.Fatalf("want err, got nil")
			}
		})
	}
}

func Test_parseQueryValues_valid(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  url.Values
	}{
		{
			name:  "Header with key-value pair as value",
			input: []string{`key1="nkey1="nval1""`, "key2=nkey2=nval2"},
			want: url.Values{
				"key1": []string{`"nkey1="nval1""`},
				"key2": []string{"nkey2=nval2"},
			},
		},
		{
			name:  "Header with normal values",
			input: []string{`key1="val1"`, "key2=val2"},
			want: url.Values{
				"key1": []string{`"val1"`},
				"key2": []string{"val2"},
			},
		},
		{
			name:  "Header with base64 string value",
			input: []string{`key="shashashaebaf43="`},
			want:  url.Values{"key": []string{`"shashashaebaf43="`}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseQueryValues(test.input)

			if err != nil {
				t.Fatalf("want: %s, got error: %s", test.want, err)
			}

			if err == nil && !reflect.DeepEqual(test.want, got) {
				t.Fatalf("want: %s, got: %s", test.want, got)
			}
		})
	}
}

func Test_parseQueryValues_invalid(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{
			name:  "Invalid query value string",
			input: []string{"invalid"},
		},
		{
			name:  "Empty query key",
			input: []string{"=bar"},
		},
		{
			name:  "Empty query value",
			input: []string{"foo="},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseQueryValues(test.input)

			if err == nil {
				t.Fatalf("want err, got nil")
			}
		})
	}
}

func Test_getRealm(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{
			header: "Bearer",
			want:   "",
		},
		{
			header: `Bearer realm="OpenFaaS API"`,
			want:   "OpenFaaS API",
		},
		{
			header: `Bearer realm="OpenFaaS API", charset="UTF-8"`,
			want:   "OpenFaaS API",
		},
		{
			header: "",
			want:   "",
		},
	}

	for _, test := range tests {
		got := getRealm(test.header)

		if test.want != got {
			t.Errorf("want: %s, got: %s", test.want, got)
		}
	}
}
