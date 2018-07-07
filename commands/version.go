// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"runtime"

	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/version"
	"github.com/spf13/cobra"
)

// GitCommit injected at build-time
var (
	shortVersion bool
)

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short-version", false, "Just print Git SHA")
	versionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")

	faasCmd.AddCommand(versionCmd)
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version [--short-version] [--gateway GATEWAY_URL]",
	Short: "Display the clients version information",
	Long: fmt.Sprintf(`The version command returns the current clients version information.

This currently consists of the GitSHA from which the client was built.
- https://github.com/openfaas/faas-cli/tree/%s`, version.GitCommit),
	Example: `  faas-cli version
  faas-cli version --short-version`,
	Run: runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	if shortVersion {
		fmt.Println(version.BuildVersion())
	} else {
		printFiglet()
		fmt.Printf("CLI commit: %s\n", version.GitCommit)
		fmt.Printf("CLI version: %s\n", version.BuildVersion())
		printServerVersions()
	}

}

func printServerVersions() {

	var services stack.Services
	var gatewayAddress string
	var yamlGateway string
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err == nil && parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress = getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))

	timeout := 2 * time.Second
	client := proxy.MakeHTTPClient(&timeout)
	response, err := client.Get(gatewayAddress + "/system/info")
	if err != nil {
		fmt.Printf("Warning could not contact gateway for version information on %s %s\n", gatewayAddress+"/system/info", err.Error())
		return
	}

	info := make(map[string]interface{})
	upstreamBody, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(upstreamBody, &info)
	if err != nil {
		fmt.Printf("Error during unmarshal of body %s %s\n", upstreamBody, err.Error())
		return
	}

	version, sha, commit := getGatewayDetails(info)
	printGatewayDetails(gatewayAddress, version, sha, commit)

	name, orchestration, sha, version := getProviderDetails(info)
	fmt.Printf("- Provider \n\tname: %s \n\torchestration: %s \n\tversion: %s \n\tsha: %s\n",
		name, orchestration, version, sha)
}

func printGatewayDetails(gatewayAddress, version, sha, commit string) {
	fmt.Printf("- Gateway \n\turi: %s", gatewayAddress)

	if version != "" {
		fmt.Printf("\n\tversion: %s", version)
		fmt.Printf("\n\tsha: %s", sha)
		fmt.Printf("\n\tcommit: %s", commit)
	}

	fmt.Println()
}

func printFiglet() {
	figletColoured := aec.BlueF.Apply(figletStr)
	if runtime.GOOS == "windows" {
		figletColoured = aec.GreenF.Apply(figletStr)
	}
	fmt.Printf(figletColoured)
}

func getGatewayDetails(m map[string]interface{}) (version, sha, commit string) {
	v := m["version"].(map[string]interface{})

	if _, ok := m["orchestration"]; !ok {
		version = v["release"].(string)
		sha = v["sha"].(string)
		commit = v["commit_message"].(string)
	}

	return
}

func getProviderDetails(m map[string]interface{}) (name, orchestration, sha, version string) {
	if k, ok := m["provider"]; ok {
		if kv, ok := k.(map[string]interface{}); ok {
			name, orchestration, sha, version = getProviderDetailsCurrent(kv)
		} else {
			name, orchestration, sha, version = getProviderDetailsLegacy(m)
		}
	}

	return
}

func getProviderDetailsLegacy(m map[string]interface{}) (name, orchestration, sha, version string) {
	name = m["provider"].(string)
	orchestration = m["orchestration"].(string)
	v := m["version"].(map[string]interface{})
	version = v["release"].(string)
	sha = v["sha"].(string)

	return
}

func getProviderDetailsCurrent(m map[string]interface{}) (name, orchestration, sha, version string) {
	v := m["version"].(map[string]interface{})
	version = v["release"].(string)
	sha = v["sha"].(string)
	name = m["provider"].(string)
	orchestration = m["orchestration"].(string)

	return
}

const figletStr = `  ___                   _____           ____
 / _ \ _ __   ___ _ __ |  ___|_ _  __ _/ ___|
| | | | '_ \ / _ \ '_ \| |_ / _` + "`" + ` |/ _` + "`" + ` \___ \
| |_| | |_) |  __/ | | |  _| (_| | (_| |___) |
 \___/| .__/ \___|_| |_|_|  \__,_|\__,_|____/
      |_|

`
