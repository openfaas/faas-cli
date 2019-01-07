package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/openfaas/faas-cli/schema"
)

//GetSecretList get secrets list
func GetSecretList(gateway string, tlsInsecure bool) ([]schema.Secret, error) {
	var results []schema.Secret

	if tlsInsecure {
		if !strings.HasPrefix(gateway, "https") {
			fmt.Println("WARNING! Communication is not secure, please consider using HTTPS. Letsencrypt.org offers free SSL/TLS certificates.")
		}
	}

	gateway = strings.TrimRight(gateway, "/")
	client := MakeHTTPClient(&defaultCommandTimeout, tlsInsecure)

	getRequest, err := http.NewRequest(http.MethodGet, gateway+"/system/secrets", nil)
	SetAuth(getRequest, gateway)

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
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", gateway)
		}

		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", gateway, jsonErr.Error())
		}

	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return results, nil
}
