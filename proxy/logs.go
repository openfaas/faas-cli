package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/openfaas/faas-provider/logs"
)

// GetLogs return stream for the logs
func (c *Client) GetLogs(ctx context.Context, params logs.Request) (<-chan logs.Message, error) {

	logRequest, err := c.newRequest(http.MethodGet, "/system/logs", nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	logRequest.URL.RawQuery = reqAsQueryValues(params).Encode()

	res, err := c.doRequest(ctx, logRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	logStream := make(chan logs.Message, 1000)
	switch res.StatusCode {
	case http.StatusOK:
		go func() {
			defer close(logStream)
			defer res.Body.Close()

			decoder := json.NewDecoder(res.Body)
			for decoder.More() {
				msg := logs.Message{}
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

func reqAsQueryValues(r logs.Request) url.Values {
	query := url.Values{}
	query.Add("name", r.Name)
	query.Add("follow", strconv.FormatBool(r.Follow))
	if r.Instance != "" {
		query.Add("instance", r.Instance)
	}

	if r.Since != nil {
		query.Add("since", r.Since.Format(time.RFC3339))
	}

	if r.Tail != 0 {
		query.Add("tail", strconv.Itoa(r.Tail))
	}

	return query
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
