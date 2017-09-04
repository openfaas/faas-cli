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

	"github.com/alexellis/faas/gateway/requests"
)

// DeployFunction call FaaS server to deploy a new function
func DeployFunction(fprocess string, gateway string, functionName string, image string, language string, replace bool, envVars map[string]string, network string, constraints []string) {

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
		fprocessTemplate = "dotnet ./bin/Debug/netcoreapp2.0/root.dll"
	}

	gateway = strings.TrimRight(gateway, "/")

	if replace {
		DeleteFunction(gateway, functionName)
	}

	// TODO: allow registry auth to be specified or read from local Docker credentials store
	req := requests.CreateFunctionRequest{
		EnvProcess:  fprocessTemplate,
		Image:       image,
		Network:     "func_functions", // todo: specify network as an override
		Service:     functionName,
		EnvVars:     envVars,
		Constraints: constraints,
	}

	reqBytes, _ := json.Marshal(&req)
	reader := bytes.NewReader(reqBytes)
	res, err := http.Post(gateway+"/system/functions", "application/json", reader)
	if err != nil {
		fmt.Println("Is FaaS deployed? Do you need to specify the -gateway flag?")
		fmt.Println(err)
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case 200, 201, 202:
		fmt.Println("Deployed.")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			fmt.Println("Server returned unexpected status code", res.StatusCode, string(bytesOut))
		}
	}

	fmt.Println(res.Status)

	deployedURL := fmt.Sprintf("URL: %s/function/%s\n", gateway, functionName)
	fmt.Println(deployedURL)
}
