// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"net/http"

	"github.com/openfaas/faas-cli/config"
)

//SetAuth sets basic auth for the given gateway
func SetAuth(req *http.Request, gateway string) {
	username, password, err := config.LookupAuthConfig(gateway)
	if err != nil {
		// no auth info found
		return
	}

	req.SetBasicAuth(username, password)
}
