// Copyright (c) OpenFaaS Ltd 2021. All rights reserved.
//
// Licensed for use with OpenFaaS Pro only
// See EULA: https://github.com/openfaas/faas/blob/master/pro/EULA.md

package commands

import (
	"testing"
)

func Test_auth(t *testing.T) {

	testCases := []struct {
		name     string
		authURL  string
		eula     bool
		clientID string
		wantErr  string
	}{
		{
			name:     "Default parameters",
			authURL:  "",
			clientID: "",
			wantErr:  "--auth-url is required and must be a valid OIDC URL",
			eula:     true,
		},
		{
			name:     "Invalid auth-url",
			authURL:  "xyz",
			clientID: "",
			wantErr:  "--auth-url is an invalid URL: xyz",
			eula:     true,
		},
		{
			name:     "Invalid eula acceptance",
			authURL:  "http://xyz",
			clientID: "id",
			wantErr:  "the auth command is only licensed for OpenFaaS Pro customers, see: https://github.com/openfaas/faas/blob/master/pro/EULA.md",
			eula:     false,
		},
		{
			name:     "Valid auth-url, invalid client-id",
			authURL:  "http://xyz",
			clientID: "",
			wantErr:  "--client-id is required",
			eula:     true,
		},
		{
			name:     "Valid auth-url and client-id",
			authURL:  "http://xyz",
			clientID: "abc",
			wantErr:  "",
			eula:     true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			err := checkValues(testCase.authURL, testCase.clientID, testCase.eula)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}

			if testCase.wantErr != gotErr {
				t.Errorf("want %q, got %q", testCase.wantErr, gotErr)
				t.Fail()
			}

		})
	}
}

func Test_makeRedirectURI_Valid(t *testing.T) {
	uri, err := makeRedirectURI("http://localhost", 31112)

	if err != nil {
		t.Fatal(err)
	}

	got := uri.String()
	want := "http://localhost:31112/oauth/callback"

	if got != want {
		t.Errorf("want %q, got %q", want, got)
		t.Fail()
	}
}

func Test_makeRedirectURI_NoSchemeIsInvalid(t *testing.T) {
	_, err := makeRedirectURI("localhost", 31112)

	if err == nil {
		t.Fatal("test should fail without a URL scheme")
	}

	got := err.Error()
	want := "a scheme is required for the URL, i.e. http://"

	if got != want {
		t.Errorf("want %q, got %q", want, got)
		t.Fail()
	}
}
