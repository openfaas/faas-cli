// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package config

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

func Test_LookupAuthConfig_WithNoConfigFile(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test1.yml"
	_, _, err := LookupAuthConfig("http://openfaas.test1")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:config file not found)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_UpdateAuthConfig_Insert(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test2.yml"
	u := "admin"
	p := "some pass"
	gatewayURL := strings.TrimRight("http://openfaas.test/", "/")
	UpdateAuthConfig(gatewayURL, u, p)

	user, pass, err := LookupAuthConfig(gatewayURL)
	if err != nil {
		t.Errorf("got error %s", err.Error())
	}

	if user != u || pass != p {
		t.Errorf("got user %s and pass %s, expected %s %s", user, pass, u, p)
	}
}

func Test_UpdateAuthConfig_Update(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test3.yml"
	u := "admin"
	p := "pass"
	gatewayURL := strings.TrimRight("http://openfaas.test/", "/")
	UpdateAuthConfig(gatewayURL, u, p)

	user, pass, err := LookupAuthConfig(gatewayURL)
	if err != nil {
		t.Errorf("got error %s", err.Error())
	}

	if user != u || pass != p {
		t.Errorf("got user %s and pass %s, expected %s %s", user, pass, u, p)
	}

	u = "admin2"
	p = "pass2"
	UpdateAuthConfig(gatewayURL, u, p)

	user, pass, err = LookupAuthConfig(gatewayURL)
	if err != nil {
		t.Errorf("got error %s", err.Error())
	}

	if user != u || pass != p {
		t.Errorf("got user %s and pass %s, expected %s %s", user, pass, u, p)
	}
}

func Test_UpdateAuthConfig_InvaidGatewayURL(t *testing.T) {
	gateway := "http//test.test"
	err := UpdateAuthConfig(gateway, "a", "b")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:invalid gateway)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_UpdateAuthConfig_EmptyGatewayURL(t *testing.T) {
	gateway := ""
	err := UpdateAuthConfig(gateway, "a", "b")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:invalid gateway)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_UpdateAuthConfig_EmptyUsername(t *testing.T) {
	err := UpdateAuthConfig("http://test.test", "", "b")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:username)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_UpdateAuthConfig_EmptyPassword(t *testing.T) {
	err := UpdateAuthConfig("http://test.test", "a", "")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:password)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_New_NoFile(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("expected to fail on empty file path")
	}
}

func Test_EnsureFile(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test6.yml"
	cfg, err := EnsureFile()
	if err != nil {
		t.Error(err.Error())
	}
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		t.Errorf("expected config at %s", cfg)
	}
}

func Test_EncodeAuth(t *testing.T) {
	token := EncodeAuth("admin", "admin")
	if token != "YWRtaW46YWRtaW4=" {
		t.Errorf("Token not matched: %s", token)
	}
}

func Test_DecodeAuth(t *testing.T) {
	u, p, err := DecodeAuth("YWRtaW46YWRtaW4=")
	if err != nil || u != "admin" || p != "admin" {
		t.Errorf("invalid base64 decoding")
	}
}

func Test_RemoveAuthConfig(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test7.yml"

	u := "admin"
	p := "pass"
	gatewayURL := strings.TrimRight("http://openfaas.test/", "/")
	UpdateAuthConfig(gatewayURL, u, p)

	gatewayURL2 := strings.TrimRight("http://openfaas.test2/", "/")
	UpdateAuthConfig(gatewayURL2, u, p)

	err := RemoveAuthConfig(gatewayURL)
	if err != nil {
		t.Errorf("got error %s", err.Error())
	}

	_, _, err = LookupAuthConfig(gatewayURL)
	if err == nil {
		t.Fatal("Error was not returned")
	}
	r := regexp.MustCompile(`(?m:no auth config found)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_RemoveAuthConfig_WithNoConfigFile(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test8.yml"
	err := RemoveAuthConfig("http://openfaas.test1")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:config file not found)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}

func Test_RemoveAuthConfig_WithUnknownGateway(t *testing.T) {
	DefaultDir, _ = ioutil.TempDir("", "faas-cli-file-test")
	DefaultFile = "test9.yml"

	u := "admin"
	p := "pass"
	gatewayURL := strings.TrimRight("http://openfaas.test/", "/")
	UpdateAuthConfig(gatewayURL, u, p)

	err := RemoveAuthConfig("http://openfaas.test1")
	if err == nil {
		t.Errorf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:gateway)`)
	if !r.MatchString(err.Error()) {
		t.Errorf("Error not matched: %s", err.Error())
	}
}
