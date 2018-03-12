// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"strings"
)

const openFaaSURLEnvironment = "OPENFAAS_URL"

func getGatewayURL(argumentURL, defaultURL, yamlURL, environmentURL string) string {
	var gatewayURL string

	if len(argumentURL) > 0 && argumentURL != defaultURL {
		gatewayURL = argumentURL
	} else if len(yamlURL) > 0 && yamlURL != defaultURL {
		gatewayURL = yamlURL
	} else if len(environmentURL) > 0 {
		gatewayURL = environmentURL
	} else {
		gatewayURL = defaultURL
	}

	gatewayURL = strings.ToLower(strings.TrimRight(gatewayURL, "/"))
	if !strings.HasPrefix(gatewayURL, "http") {
		gatewayURL = fmt.Sprintf("http://%s", gatewayURL)
	}

	return gatewayURL
}
