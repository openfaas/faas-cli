// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/openfaas/faas/gateway/requests"
)

// DeleteFunction delete a function from the OpenFaaS server
func (c *Client) DeleteFunction(ctx context.Context, functionName string, namespace string) error {
	var err error
	delReq := requests.DeleteFunctionRequest{FunctionName: functionName}
	reqBytes, _ := json.Marshal(&delReq)
	reader := bytes.NewReader(reqBytes)
	deleteEndpoint := "/system/functions"

	query := url.Values{}
	if len(namespace) > 0 {
		query.Add("namespace", namespace)
	}

	req, err := c.newRequest(http.MethodDelete, deleteEndpoint, query, reader)
	if err != nil {
		fmt.Println(err)
		return err
	}

	res, err := c.doRequest(ctx, req)
	if err != nil {
		fmt.Printf("Error removing existing function: %s, gateway=%s, functionName=%s\n",
			err.Error(), c.GatewayURL.String(), functionName)
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		fmt.Println("Removing old function.")
	case http.StatusNotFound:
		err = fmt.Errorf("No existing function to remove")
	case http.StatusUnauthorized:
		err = fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		var bodyReadErr error
		bytesOut, bodyReadErr := io.ReadAll(res.Body)
		if bodyReadErr != nil {
			err = bodyReadErr
		} else {
			err = fmt.Errorf("Server returned unexpected status code %d %s", res.StatusCode, string(bytesOut))
		}
	}

	return err
}
