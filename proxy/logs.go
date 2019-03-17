package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/openfaas/faas-cli/schema"
)

// GetLogs list deployed functions
func GetLogs(gateway string, tlsInsecure bool, params schema.LogRequest) (<-chan schema.LogMessage, error) {

	gateway = strings.TrimRight(gateway, "/")
	// replace with a client that allows keep alive, Default?
	client := makeStreamingHTTPClient(tlsInsecure)

	logRequest, err := http.NewRequest(http.MethodGet, gateway+"/system/logs", nil)
	SetAuth(logRequest, gateway)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	logRequest.URL.RawQuery = params.AsQueryValues().Encode()

	res, err := client.Do(logRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", gateway)
	}

	logStream := make(chan schema.LogMessage, 1000)
	switch res.StatusCode {
	case http.StatusOK:
		go func() {
			defer close(logStream)
			defer res.Body.Close()

			decoder := json.NewDecoder(res.Body)
			for decoder.More() {
				msg := schema.LogMessage{}
				err := decoder.Decode(&msg)
				if err != nil {
					log.Printf("cannot parse log results: %s\n", err.Error())
					return
				}
				logStream <- msg
			}
		}()
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return logStream, nil
}

func makeStreamingHTTPClient(tlsInsecure bool) http.Client {
	client := http.Client{}

	if tlsInsecure {
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}

		if tlsInsecure {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: tlsInsecure}
		}

		client.Transport = tr
	}

	return client
}
