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
	"time"

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas/gateway/requests"
)

// FunctionResourceRequest defines a request to set function resources
type FunctionResourceRequest struct {
	Limits   *stack.FunctionResources
	Requests *stack.FunctionResources
}

// DeployFunction call FaaS server to deploy a new function
func DeployFunction(fprocess string, gateway string, functionName string, image string,
	language string, replace bool, envVars map[string]string, network string,
	constraints []string, update bool, secrets []string, labels map[string]string, functionResourceRequest1 FunctionResourceRequest) {

	// Need to alter Gateway to allow nil/empty string as fprocess, to avoid this repetition.
	var fprocessTemplate string
	if len(fprocess) > 0 {
		fprocessTemplate = fprocess
	}

	gateway = strings.TrimRight(gateway, "/")

	if replace {
		DeleteFunction(gateway, functionName)
	}

	req := requests.CreateFunctionRequest{
		EnvProcess:  fprocessTemplate,
		Image:       image,
		Network:     network,
		Service:     functionName,
		EnvVars:     envVars,
		Constraints: constraints,
		Secrets:     secrets, // TODO: allow registry auth to be specified or read from local Docker credentials store
		Labels:      &labels,
	}

	req.Limits = &requests.FunctionResources{}
	req.Requests = &requests.FunctionResources{}

	if functionResourceRequest1.Limits != nil && len(functionResourceRequest1.Limits.Memory) > 0 {
		req.Limits.Memory = functionResourceRequest1.Limits.Memory
	}

	if functionResourceRequest1.Requests != nil && len(functionResourceRequest1.Requests.Memory) > 0 {
		req.Requests.Memory = functionResourceRequest1.Requests.Memory
	}

	if functionResourceRequest1.Limits != nil && len(functionResourceRequest1.Limits.CPU) > 0 {
		req.Limits.CPU = functionResourceRequest1.Limits.CPU
	}

	if functionResourceRequest1.Requests != nil && len(functionResourceRequest1.Requests.CPU) > 0 {
		req.Requests.CPU = functionResourceRequest1.Requests.CPU
	}

	reqBytes, _ := json.Marshal(&req)
	reader := bytes.NewReader(reqBytes)
	var request *http.Request

	timeout := 60 * time.Second
	client := MakeHTTPClient(&timeout)

	method := http.MethodPost
	// "application/json"
	if update {
		method = http.MethodPut
	}

	var err error
	request, err = http.NewRequest(method, gateway+"/system/functions", reader)
	SetAuth(request, gateway)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(request)
	if err != nil {
		fmt.Println("Is FaaS deployed? Do you need to specify the --gateway flag?")
		fmt.Println(err)
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		if update {
			fmt.Println("Updated.")
		} else {
			fmt.Println("Deployed.")
		}

		deployedURL := fmt.Sprintf("URL: %s/function/%s\n", gateway, functionName)
		fmt.Println(deployedURL)
	case http.StatusUnauthorized:
		fmt.Println("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			fmt.Printf("Unexpected status: %d, message: %s\n", res.StatusCode, string(bytesOut))
		}
	}

	fmt.Println(res.Status)
}
