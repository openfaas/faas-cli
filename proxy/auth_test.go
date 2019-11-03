// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"net/http"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/openfaas/faas-cli/config"
)

func Test_SetAuth_AuthorizationHeader(t *testing.T) {
	//setup store
	config.DefaultDir, _ = ioutil.TempDir("", "faas-cli-auth-test")
	config.DefaultFile = "authtest1.yml"
	basicAuthURL := strings.TrimRight("http://openfaas.test/", "/")
	openURL := "http://openfaas.test/"
	token := config.EncodeAuth("Aladdin", "open sesame")
	config.UpdateAuthConfig(basicAuthURL, token, config.BasicAuthType)

	req, _ := http.NewRequest("GET", openURL, nil)
	SetAuth(req, basicAuthURL)
	header := req.Header.Get("Authorization")
	expected := "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="
	if header != expected {
		t.Errorf("got header %q, want %q", header, expected)
	}
}

func Test_SetAuth_SkipAuthorization(t *testing.T) {
	//setup store
	config.DefaultDir, _ = ioutil.TempDir("", "faas-cli-auth-test")
	config.DefaultFile = "authtest2.yml"
	basicAuthURL := strings.TrimRight("http://openfaas.test/", "/")
	openURL := "http://openfaas.test2/"
	token := config.EncodeAuth("Aladdin", "open sesame")
	config.UpdateAuthConfig(basicAuthURL, token, config.BasicAuthType)

	req, _ := http.NewRequest("GET", openURL, nil)
	SetAuth(req, "http://openfaas.test2")
	header := req.Header.Get("Authorization")
	if header != "" {
		t.Errorf("got header %q, want none", header)
	}
}

func Test_SetAuth_Oauth2(t *testing.T) {
	//setup store
	config.DefaultDir, _ = ioutil.TempDir("", "faas-cli-auth-test")
	config.DefaultFile = "authtest2.yml"
	basicAuthURL := strings.TrimRight("http://openfaas.ouath.test/", "/")
	token := "somebase64string"
	config.UpdateAuthConfig(basicAuthURL, token, config.Oauth2AuthType)

	req, _ := http.NewRequest("GET", basicAuthURL, nil)
	SetAuth(req, basicAuthURL)
	header := req.Header.Get("Authorization")
	expectedValue := "Bearer " + token
	if header != expectedValue {
		t.Errorf("got header %q, want %q", header, expectedValue)
	}
}

func Test_SetAuth_BasicAuth(t *testing.T) {
	//setup store
	config.DefaultDir, _ = ioutil.TempDir("", "faas-cli-auth-test")
	config.DefaultFile = "authtest2.yml"
	basicAuthURL := strings.TrimRight("http://openfaas.basic-auth.test/", "/")
	token := config.EncodeAuth("username", "password")
	config.UpdateAuthConfig(basicAuthURL, token, config.BasicAuthType)

	req, _ := http.NewRequest("GET", basicAuthURL, nil)
	SetAuth(req, basicAuthURL)
	header := req.Header.Get("Authorization")
	expectedValue := "Basic " + token
	if header != expectedValue {
		t.Errorf("got header %q, want %q", header, expectedValue)
	}
}

func Test_SetAuth_SkipAuthorizationHeader(t *testing.T) {
	//setup store
	config.DefaultDir, _ = ioutil.TempDir("", "faas-cli-auth-test")
	config.DefaultFile = "authtest2.yml"
	basicAuthURL := strings.TrimRight("http://openfaas.test/", "/")
	openURL := "http://openfaas.test2/"
	token := config.EncodeAuth("Aladdin", "open sesame")
	config.UpdateAuthConfig(basicAuthURL, token, config.Oauth2AuthType)

	req, _ := http.NewRequest("GET", openURL, nil)
	SetAuth(req, openURL)
	header := req.Header.Get("Authorization")
	if header != "" {
		t.Errorf("got header %q, want none", header)
	}
}
