package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v2 "github.com/openfaas/faas-cli/schema/store/v2"
)

type StoreResult struct {
	Version   string             `json:"version"`
	Functions []v2.StoreFunction `json:"functions"`
}

// FunctionStoreList returns functions from a store URL
func FunctionStoreList(store string) ([]v2.StoreFunction, error) {

	var storeResults StoreResult

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
		bytesOut, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS store at URL: %s", store)
		}

		jsonErr := json.Unmarshal(bytesOut, &storeResults)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS store at URL: %s\n%s", store, jsonErr.Error())
		}
	default:
		bytesOut, err := io.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return storeResults.Functions, nil
}
