package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/schema"
)

// FunctionStoreList returns functions from a store URL
func FunctionStoreList(store string) ([]schema.StoreItem, error) {
	var results []schema.StoreItem

	store = strings.TrimRight(store, "/")

	timeout := 60 * time.Second
	tlsInsecure := false

	client := MakeHTTPClient(&timeout, tlsInsecure)

	res, err := client.Get(store)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS store at URL: %s", store)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS store at URL: %s", store)
		}

		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS store at URL: %s\n%s", store, jsonErr.Error())
		}
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return results, nil
}
