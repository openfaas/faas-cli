package proxy

import (
	"fmt"
	"net/url"
	"path"
)

const (
	systemPath     = "/system/functions"
	functionPath   = "/system/function"
	namespacesPath = "/system/namespaces"
	namespaceKey   = "namespace"
)

func createSystemEndpoint(gateway, namespace string) (string, error) {
	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		return "", fmt.Errorf("invalid gateway URL: %s", err.Error())
	}
	gatewayURL.Path = path.Join(gatewayURL.Path, systemPath)
	if len(namespace) > 0 {
		q := gatewayURL.Query()
		q.Set("namespace", namespace)
		gatewayURL.RawQuery = q.Encode()
	}
	return gatewayURL.String(), nil
}

func createFunctionEndpoint(gateway, functionName, namespace string) (string, error) {
	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		return "", fmt.Errorf("invalid gateway URL: %s", err.Error())
	}
	gatewayURL.Path = path.Join(gatewayURL.Path, functionPath, functionName)
	if len(namespace) > 0 {
		q := gatewayURL.Query()
		q.Set("namespace", namespace)
		gatewayURL.RawQuery = q.Encode()
	}
	return gatewayURL.String(), nil
}

func createNamespacesEndpoint(gateway string) (string, error) {
	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		return "", fmt.Errorf("invalid gateway URL: %s", err.Error())
	}
	gatewayURL.Path = path.Join(gatewayURL.Path, namespacesPath)
	return gatewayURL.String(), nil
}
