// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/stack"

	types "github.com/openfaas/faas-provider/types"
)

var (
	defaultCommandTimeout = 60 * time.Second
)

// FunctionResourceRequest defines a request to set function resources
type FunctionResourceRequest struct {
	Limits   *stack.FunctionResources
	Requests *stack.FunctionResources
}

// DeployFunctionSpec defines the spec used when deploying a function
type DeployFunctionSpec struct {
	FProcess                string
	Gateway                 string
	FunctionName            string
	Image                   string
	RegistryAuth            string
	Language                string
	Replace                 bool
	EnvVars                 map[string]string
	Network                 string
	Constraints             []string
	Update                  bool
	Secrets                 []string
	Labels                  map[string]string
	Annotations             map[string]string
	FunctionResourceRequest FunctionResourceRequest
	ReadOnlyRootFilesystem  bool
	TLSInsecure             bool
	Token                   string
	Namespace               string
}

// DeployFunction first tries to deploy a function and if it exists will then attempt
// a rolling update. Warnings are suppressed for the second API call (if required.)
func (c *Client) DeployFunction(context context.Context, spec *DeployFunctionSpec) int {

	rollingUpdateInfo := fmt.Sprintf("Function %s already exists, attempting rolling-update.", spec.FunctionName)
	warnInsecureGateway := true
	statusCode, deployOutput := c.Deploy(context, spec, spec.Update, warnInsecureGateway)

	warnInsecureGateway = false
	if spec.Update == true && statusCode == http.StatusNotFound {
		// Re-run the function with update=false

		statusCode, deployOutput = c.Deploy(context, spec, false, warnInsecureGateway)
	} else if statusCode == http.StatusOK {
		fmt.Println(rollingUpdateInfo)
	}
	fmt.Println()
	fmt.Println(deployOutput)
	return statusCode
}

// Deploy a function to an OpenFaaS gateway over REST
func (c *Client) Deploy(context context.Context, spec *DeployFunctionSpec, update bool, warnInsecureGateway bool) (int, string) {

	var deployOutput string
	// Need to alter Gateway to allow nil/empty string as fprocess, to avoid this repetition.
	var fprocessTemplate string
	if len(spec.FProcess) > 0 {
		fprocessTemplate = spec.FProcess
	}

	gateway := strings.TrimRight(spec.Gateway, "/")

	if spec.Replace {
		DeleteFunction(gateway, spec.FunctionName, spec.TLSInsecure, "")
	}

	req := types.FunctionDeployment{
		EnvProcess:             fprocessTemplate,
		Image:                  spec.Image,
		RegistryAuth:           spec.RegistryAuth,
		Network:                spec.Network,
		Service:                spec.FunctionName,
		EnvVars:                spec.EnvVars,
		Constraints:            spec.Constraints,
		Secrets:                spec.Secrets,
		Labels:                 &spec.Labels,
		Annotations:            &spec.Annotations,
		ReadOnlyRootFilesystem: spec.ReadOnlyRootFilesystem,
		Namespace:              spec.Namespace,
	}

	hasLimits := false
	req.Limits = &types.FunctionResources{}
	if spec.FunctionResourceRequest.Limits != nil && len(spec.FunctionResourceRequest.Limits.Memory) > 0 {
		hasLimits = true
		req.Limits.Memory = spec.FunctionResourceRequest.Limits.Memory
	}
	if spec.FunctionResourceRequest.Limits != nil && len(spec.FunctionResourceRequest.Limits.CPU) > 0 {
		hasLimits = true
		req.Limits.CPU = spec.FunctionResourceRequest.Limits.CPU
	}
	if !hasLimits {
		req.Limits = nil
	}

	hasRequests := false
	req.Requests = &types.FunctionResources{}
	if spec.FunctionResourceRequest.Requests != nil && len(spec.FunctionResourceRequest.Requests.Memory) > 0 {
		hasRequests = true
		req.Requests.Memory = spec.FunctionResourceRequest.Requests.Memory
	}
	if spec.FunctionResourceRequest.Requests != nil && len(spec.FunctionResourceRequest.Requests.CPU) > 0 {
		hasRequests = true
		req.Requests.CPU = spec.FunctionResourceRequest.Requests.CPU
	}

	if !hasRequests {
		req.Requests = nil
	}

	reqBytes, _ := json.Marshal(&req)
	reader := bytes.NewReader(reqBytes)
	var request *http.Request

	client := MakeHTTPClient(&defaultCommandTimeout, spec.TLSInsecure)

	method := http.MethodPost
	// "application/json"
	if update {
		method = http.MethodPut
	}

	var err error
	request, err = http.NewRequest(method, gateway+"/system/functions", reader)
	if len(spec.Token) > 0 {
		SetToken(request, spec.Token)
	} else {
		SetAuth(request, gateway)
	}

	if err != nil {
		deployOutput += fmt.Sprintln(err)
		return http.StatusInternalServerError, deployOutput
	}

	res, err := client.Do(request.WithContext(context))
	if err != nil {
		deployOutput += fmt.Sprintln("Is OpenFaaS deployed? Do you need to specify the --gateway flag?")
		deployOutput += fmt.Sprintln(err)
		return http.StatusInternalServerError, deployOutput
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		deployOutput += fmt.Sprintf("Deployed. %s.\n", res.Status)

		deployedURL := fmt.Sprintf("URL: %s/function/%s", gateway, spec.FunctionName)
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
