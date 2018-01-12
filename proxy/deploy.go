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

func DeployFunction(fprocess string, gateway string, functionName string, image string,
	language string, replace bool, envVars map[string]string, network string,
	constraints []string, update bool, secrets []string, labels map[string]string,
	functionResourceRequest1 FunctionResourceRequest) {

	rollingUpdateInfo := fmt.Sprintf("Function %s already exists, attempting rolling-update.", functionName)
	statusCode, deployOutput := Deploy(fprocess, gateway, functionName, image, language, replace, envVars, network, constraints, update, secrets, labels, functionResourceRequest1)

	if update == true && statusCode == http.StatusNotFound {
		// Re-run the function with update=false
		_, deployOutput = Deploy(fprocess, gateway, functionName, image, language, replace, envVars, network, constraints, false, secrets, labels, functionResourceRequest1)
	} else if statusCode == http.StatusOK {
		fmt.Println(rollingUpdateInfo)
	}
	fmt.Println()
	fmt.Println(deployOutput)
}

func Deploy(fprocess string, gateway string, functionName string, image string,
	language string, replace bool, envVars map[string]string, network string,
	constraints []string, update bool, secrets []string, labels map[string]string,
	functionResourceRequest1 FunctionResourceRequest) (int, string) {

	var deployOutput string
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

	hasLimits := false
	req.Limits = &requests.FunctionResources{}
	if functionResourceRequest1.Limits != nil && len(functionResourceRequest1.Limits.Memory) > 0 {
		hasLimits = true
		req.Limits.Memory = functionResourceRequest1.Limits.Memory
	}
	if functionResourceRequest1.Limits != nil && len(functionResourceRequest1.Limits.CPU) > 0 {
		hasLimits = true
		req.Limits.CPU = functionResourceRequest1.Limits.CPU
	}
	if !hasLimits {
		req.Limits = nil
	}

	hasRequests := false
	req.Requests = &requests.FunctionResources{}
	if functionResourceRequest1.Requests != nil && len(functionResourceRequest1.Requests.Memory) > 0 {
		hasRequests = true
		req.Requests.Memory = functionResourceRequest1.Requests.Memory
	}
	if functionResourceRequest1.Requests != nil && len(functionResourceRequest1.Requests.CPU) > 0 {
		hasRequests = true
		req.Requests.CPU = functionResourceRequest1.Requests.CPU
	}

	if !hasRequests {
		req.Requests = nil
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
		deployOutput += fmt.Sprintln(err)
		return http.StatusInternalServerError, deployOutput
	}

	res, err := client.Do(request)
	if err != nil {
		deployOutput += fmt.Sprintln("Is FaaS deployed? Do you need to specify the --gateway flag?")
		deployOutput += fmt.Sprintln(err)
		return http.StatusInternalServerError, deployOutput
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		deployOutput += fmt.Sprintf("Deployed. %s.\n", res.Status)

		deployedURL := fmt.Sprintf("URL: %s/function/%s", gateway, functionName)
		deployOutput += fmt.Sprintln(deployedURL)
	case http.StatusUnauthorized:
		deployOutput += fmt.Sprintln("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
		/*
			case http.StatusNotFound:
				if replace && !update {
					deployOutput += fmt.Sprintln("Could not delete-and-replace function because it is not found (404)")
				}
		*/
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			deployOutput += fmt.Sprintf("Unexpected status: %d, message: %s\n", res.StatusCode, string(bytesOut))
		}
	}

	return res.StatusCode, deployOutput
}
