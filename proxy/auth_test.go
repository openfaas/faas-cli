// Copyright (c) OpenFaaS Project 2017. All rights reserved.
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
	config.UpdateAuthConfig(basicAuthURL, "Aladdin", "open sesame")

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
	config.UpdateAuthConfig(basicAuthURL, "Aladdin", "open sesame")

	req, _ := http.NewRequest("GET", openURL, nil)
	SetAuth(req, "http://openfaas.test2")
	header := req.Header.Get("Authorization")
	if header != "" {
		t.Errorf("got header %q, want none", header)
	}
}
