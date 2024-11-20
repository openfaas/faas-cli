package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
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

	err := ReadJSON(context.TODO(), store, &storeResults)
	if err != nil {
		return nil, fmt.Errorf("cannot read result from OpenFaaS store at URL: %s", store)
	}

	return storeResults.Functions, nil
}

// ReadJSON reads a JSON file from a URL or local file
func ReadJSON(ctx context.Context, location string, dest interface{}) error {
	var body io.ReadCloser
	var err error

	timeout := 60 * time.Second
	tlsInsecure := false

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	scheme := determineScheme(location)
	switch scheme {
	case "http", "https":
		client := MakeHTTPClient(&timeout, tlsInsecure)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		if err != nil {
			return fmt.Errorf("cannot create request to: %s", location)
		}

		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("cannot connect to: %s", location)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned unexpected status code: %d", res.StatusCode)
		}

		body = res.Body
	case "file":
		location, err = expandTilde(location)
		if err != nil {
			return err
		}

		body, err = os.Open(location)
		if err != nil {
			return fmt.Errorf("cannot read file: %s", location)
		}

	// Add more schemes such as s3:// or gs://
	default:
		return fmt.Errorf("unsupported scheme: %s", scheme)
	}

	if body != nil {
		defer body.Close()
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("cannot read data from: %s", location)
	}

	return json.Unmarshal(data, dest)
}

func determineScheme(location string) string {
	location = strings.ToLower(location)
	if strings.HasPrefix(location, "http://") {
		return "http"
	}
	if strings.HasPrefix(location, "https://") {
		return "https"
	}
	return "file"
}

// expandTilde expands a path with a leading tilde to the home directory
func expandTilde(location string) (string, error) {
	if !strings.HasPrefix(location, "~") {
		return location, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return strings.Replace(location, "~", usr.HomeDir, 1), nil
}
