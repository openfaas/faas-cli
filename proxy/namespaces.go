package proxy

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// ListNamespaces lists available function namespaces
func ListNamespaces(gateway string, tlsInsecure bool) ([]string, error) {
	return ListNamespacesToken(gateway, tlsInsecure, "")
}

// ListNamespacesToken lists available function namespaces with a token as auth
func ListNamespacesToken(gateway string, tlsInsecure bool, token string) ([]string, error) {
	var namespaces []string

	gateway = strings.TrimRight(gateway, "/")
	client := MakeHTTPClient(&defaultCommandTimeout, tlsInsecure)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	getEndpoint, err := createNamespacesEndpoint(gateway)
	if err != nil {
		return namespaces, err
	}

	getRequest, err := http.NewRequest(http.MethodGet, getEndpoint, nil)

	if len(token) > 0 {
		SetToken(getRequest, token)
	} else {
		SetAuth(getRequest, gateway)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	res, err := client.Do(getRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read namespaces from OpenFaaS on URL: %s", gateway)
		}
		jsonErr := json.Unmarshal(bytesOut, &namespaces)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse namespaces from OpenFaaS on URL: %s\n%s", gateway, jsonErr.Error())
		}
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return namespaces, nil
}
