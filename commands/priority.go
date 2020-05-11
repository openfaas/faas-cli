// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"strings"
)

const (
	openFaaSURLEnvironment      = "OPENFAAS_URL"
	templateURLEnvironment      = "OPENFAAS_TEMPLATE_URL"
	templateStoreURLEnvironment = "OPENFAAS_TEMPLATE_STORE_URL"
)

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

func getTemplateURL(argumentURL, environmentURL, defaultURL string) string {
	var templateURL string

	if len(argumentURL) > 0 && argumentURL != defaultURL {
		templateURL = argumentURL
	} else if len(environmentURL) > 0 {
		templateURL = environmentURL
	} else {
		templateURL = defaultURL
	}

	return templateURL
}

func getTemplateStoreURL(argumentURL, environmentURL, defaultURL string) string {
	if argumentURL != defaultURL {
		return argumentURL
	} else if len(environmentURL) > 0 {
		return environmentURL
	} else {
		return defaultURL
	}
}

func getNamespace(flagNamespace, stackNamespace string) string {
	// If the namespace flag is passed use it
	if len(flagNamespace) > 0 {
		return flagNamespace
	}

	// if both the namespace flag in stack.yaml and the namespace flag are ommitted
	// return the defaultNamespace (openfaas-fn)
	if len(stackNamespace) == 0 && len(flagNamespace) == 0 {
		return defaultFunctionNamespace
	}

	// Else return the namespace in stack.yaml
	return stackNamespace
}
