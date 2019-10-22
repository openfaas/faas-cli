// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"net/http"

	"github.com/openfaas/faas-cli/config"
)

//SetAuth sets basic auth for the given gateway
func SetAuth(req *http.Request, gateway string) {
	authConfig, err := config.LookupAuthConfig(gateway)
	if err != nil {
		// no auth info found
		return
	}
	username, password, err := config.DecodeAuth(authConfig.Token)
	if err != nil {
		// no auth info found
		return
	}
	req.SetBasicAuth(username, password)
}

//SetToken sets authentication token
func SetToken(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

//SetBasicAuth set basic authentication
func SetBasicAuth(req *http.Request, authConfig config.AuthConfig) {
	username, password, err := config.DecodeAuth(authConfig.Token)
	if err != nil {
		// no auth info found
		return
	}
	req.SetBasicAuth(username, password)
}

//SetOauth2 set oauth2 token
func SetOauth2(req *http.Request, authConfig config.AuthConfig) {
	SetToken(req, authConfig.Token)
}

//AddAuth add authentication
func AddAuth(req *http.Request, gateway string) {
	authConfig, err := config.LookupAuthConfig(gateway)
	if err != nil {
		// no auth info found
		return
	}

	if authConfig.Auth == config.BasicAuthType {
		SetBasicAuth(req, authConfig)
	} else if authConfig.Auth == config.Oauth2AuthType {
		SetOauth2(req, authConfig)
	}
}
