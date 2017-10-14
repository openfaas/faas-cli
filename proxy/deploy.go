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

	"github.com/openfaas/faas/gateway/requests"
)

// DeployFunction call FaaS server to deploy a new function
func DeployFunction(fprocess string, gateway string, functionName string, image string,
	language string, replace bool, envVars map[string]string, network string,
	constraints []string, update bool, secrets []string, labels map[string]string) {

	// Need to alter Gateway to allow nil/empty string as fprocess, to avoid this repetition.
	var fprocessTemplate string
	if len(fprocess) > 0 {
		fprocessTemplate = fprocess
	} else if language == "python" {
		fprocessTemplate = "python index.py"
	} else if language == "node" {
		fprocessTemplate = "node index.js"
	} else if language == "ruby" {
		fprocessTemplate = "ruby index.rb"
	} else if language == "csharp" {
		fprocessTemplate = "dotnet ./root.dll"
	} else if language == "python3" {
		fprocessTemplate = "python3 index.py"
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

	reqBytes, _ := json.Marshal(&req)
	reader := bytes.NewReader(reqBytes)
	var request *http.Request

	timeout := 120 * time.Second
	client := MakeHTTPClient(&timeout)

	method := "POST"
	// "application/json"
	if update {
		method = "PUT"
	}

	request, _ = http.NewRequest(method, gateway+"/system/functions", reader)
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
	case 200, 201, 202:
		if update {
			fmt.Println("Updated.")
		} else {
			fmt.Println("Deployed.")
		}

		deployedURL := fmt.Sprintf("URL: %s/function/%s\n", gateway, functionName)
		fmt.Println(deployedURL)
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			fmt.Printf("Unexpected status: %d, message: %s\n", res.StatusCode, string(bytesOut))
		}
	}

	fmt.Println(res.Status)
}
