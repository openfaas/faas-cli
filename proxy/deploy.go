package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
)

// DeployFunction call FaaS server to deploy a new function
func DeployFunction(fprocess string, gateway string, functionName string, image string, language string, replace bool, envVars map[string]string, network string) {

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

	if replace {
		DeleteFunction(gateway, functionName)
	}

	// TODO: allow registry auth to be specified or read from local Docker credentials store
	req := requests.CreateFunctionRequest{
		EnvProcess: fprocessTemplate,
		Image:      image,
		Network:    "func_functions", // todo: specify network as an override
		Service:    functionName,
		EnvVars:    envVars,
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
