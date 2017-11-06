// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/openfaas/faas/gateway/requests"
)

// DeleteFunction delete a function from the FaaS server
func DeleteFunction(gateway string, functionName string) error {
	gateway = strings.TrimRight(gateway, "/")
	delReq := requests.DeleteFunctionRequest{FunctionName: functionName}
	reqBytes, _ := json.Marshal(&delReq)
	reader := bytes.NewReader(reqBytes)

	c := http.Client{}
	req, err := http.NewRequest("DELETE", gateway+"/system/functions", reader)
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	delRes, delErr := c.Do(req)

	if delErr != nil {
		fmt.Printf("Error removing existing function: %s, gateway=%s, functionName=%s\n", delErr.Error(), gateway, functionName)
		return delErr
	}

	if delRes.Body != nil {
		defer delRes.Body.Close()
	}

	switch delRes.StatusCode {
	case 200, 201, 202:
		fmt.Println("Removing old function.")
	case 404:
		fmt.Println("No existing function to remove")
	default:
		var bodyReadErr error
		bytesOut, bodyReadErr := ioutil.ReadAll(delRes.Body)
		if bodyReadErr != nil {
			err = bodyReadErr
		} else {
			err = fmt.Errorf("server returned unexpected status code %d %s", delRes.StatusCode, string(bytesOut))
			fmt.Println("Server returned unexpected status code", delRes.StatusCode, string(bytesOut))
		}
	}

	return err
}
