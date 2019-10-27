package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	types "github.com/openfaas/faas-provider/types"
)

const (
	secretEndpoint = "/system/secrets"
)

// GetSecretList get secrets list
func (c *Client) GetSecretList(ctx context.Context, namespace string) ([]types.Secret, error) {
	var (
		results    []types.Secret
		err        error
		secretPath = secretEndpoint
	)

	if len(namespace) > 0 {
		secretPath, err = addQueryParams(secretPath, map[string]string{namespaceKey: namespace})
	}

	getRequest, err := c.newRequest(http.MethodGet, secretPath, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	res, err := c.doRequest(ctx, getRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusAccepted:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", c.GatewayURL.String())
		}

		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", c.GatewayURL.String(), jsonErr.Error())
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

// UpdateSecret update a secret via the OpenFaaS API by name
func (c *Client) UpdateSecret(ctx context.Context, secret types.Secret) (int, string) {
	var output string
	reqBytes, _ := json.Marshal(&secret)

	putRequest, err := c.newRequest(http.MethodPut, secretEndpoint, bytes.NewBuffer(reqBytes))

	if err != nil {
		output += fmt.Sprintf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
		return http.StatusInternalServerError, output
	}

	res, err := c.doRequest(ctx, putRequest)
	if err != nil {
		output += fmt.Sprintf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
		return http.StatusInternalServerError, output
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		output += fmt.Sprintf("Updated: %s\n", res.Status)
		break

	case http.StatusNotFound:
		output += fmt.Sprintf("unable to find secret: %s", secret.Name)

	case http.StatusUnauthorized:
		output += fmt.Sprintf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			output += fmt.Sprintf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return res.StatusCode, output
}

// RemoveSecret remove a secret via the OpenFaaS API by name
func (c *Client) RemoveSecret(ctx context.Context, secret types.Secret) error {
	body, _ := json.Marshal(secret)
	req, err := c.newRequest(http.MethodDelete, secretEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	res, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		break
	case http.StatusNotFound:
		return fmt.Errorf("unable to find secret: %s", secret.Name)
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return nil
}

// CreateSecret create secret
func (c *Client) CreateSecret(ctx context.Context, secret types.Secret) (int, string) {
	var output string
	reqBytes, _ := json.Marshal(&secret)
	reader := bytes.NewReader(reqBytes)

	request, err := c.newRequest(http.MethodPost, secretEndpoint, reader)

	if err != nil {
		output += fmt.Sprintf("cannot connect to OpenFaaS on URL: %s\n", c.GatewayURL.String())
		return http.StatusInternalServerError, output
	}

	res, err := c.doRequest(ctx, request)
	if err != nil {
		output += fmt.Sprintf("cannot connect to OpenFaaS on URL: %s\n", c.GatewayURL.String())
		return http.StatusInternalServerError, output
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		output += fmt.Sprintf("Created: %s\n", res.Status)

	case http.StatusUnauthorized:
		output += fmt.Sprintln("unauthorized access, run \"faas-cli login\" to setup authentication for this server")

	case http.StatusConflict:
		output += fmt.Sprintf("secret with the name %q already exists\n", secret.Name)

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			output += fmt.Sprintf("server returned unexpected status code: %d - %s\n", res.StatusCode, string(bytesOut))
		}
	}

	return res.StatusCode, output
}
