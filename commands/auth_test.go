// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"
)

func Test_auth(t *testing.T) {

	testCases := []struct {
		name     string
		authURL  string
		clientID string
		wantErr  string
	}{
		{
			name:     "Default parameters",
			authURL:  "",
			clientID: "",
			wantErr:  "--auth-url is required and must be a valid OIDC /authorize URL",
		},
		{
			name:     "Invalid auth-url",
			authURL:  "xyz",
			clientID: "",
			wantErr:  "--auth-url is an invalid URL: xyz",
		},
		{
			name:     "Valid auth-url, invalid client-id",
			authURL:  "http://xyz",
			clientID: "",
			wantErr:  "--client-id is required",
		},
		{
			name:     "Valid auth-url and client-id",
			authURL:  "http://xyz",
			clientID: "abc",
			wantErr:  "",
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			err := checkValues(testCase.authURL, testCase.clientID)
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
