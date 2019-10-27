// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"os"

	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/version"
	"github.com/spf13/cobra"
)

// GitCommit injected at build-time
var (
	shortVersion bool
	warnUpdate   bool
)

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short-version", false, "Just print Git SHA")
	versionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	versionCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	versionCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")

	versionCmd.Flags().BoolVar(&warnUpdate, "warn-update", true, "Check for new version and warn about updating")

	versionCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
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
	RunE: runVersionE,
}

func runVersionE(cmd *cobra.Command, args []string) error {
	releases := "https://github.com/openfaas/faas-cli/releases/latest"

	if shortVersion {
		fmt.Println(version.BuildVersion())

	} else {
		printLogo()
		fmt.Printf(`CLI:
 commit:  %s
 version: %s
`, version.GitCommit, version.BuildVersion())
		printServerVersions()
	}

	if warnUpdate {
		version := version.Version
		latest, err := findRelease(releases)
		if err != nil {
			return fmt.Errorf("unable to find latest version online error: %s", err.Error())
		}

		if version != "" && version != latest {
			fmt.Printf("Your faas-cli version (%s) may be out of date. Version: %s is now available on GitHub.\n", version, latest)
		}
	}

	return nil
}

func printServerVersions() {

	var services stack.Services
	var gatewayAddress string
	var yamlGateway string
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err == nil && parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress = getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))

	versionTimeout := 5 * time.Second
	cliAuth := NewCLIAuth(token, gatewayAddress)
	transport := GetDefaultCLITransport(tlsInsecure, &versionTimeout)
	cliClient := proxy.NewClient(cliAuth, gatewayAddress, transport, &versionTimeout)
	info, err := cliClient.GetSystemInfo(context.Background())
	if err != nil {
		return
	}

	version, sha, commit := getGatewayDetails(info)

	printGatewayDetails(gatewayAddress, version, sha, commit)

	name, orchestration, sha, version := getProviderDetails(info)
	fmt.Printf(`
Provider
 name:          %s
 orchestration: %s
 version:       %s 
 sha:           %s
`, name, orchestration, version, sha)
}

func printGatewayDetails(gatewayAddress, version, sha, commit string) {
	fmt.Printf(`
Gateway
 uri:     %s`, gatewayAddress)

	if version != "" {
		fmt.Printf(`
 version: %s
 sha:     %s
 commit:  %s
`, version, sha, commit)
	}

	fmt.Println()
}

// printLogo prints an ASCII logo, which was generated with figlet
func printLogo() {
	figletColoured := aec.BlueF.Apply(figletStr)
	if runtime.GOOS == "windows" {
		figletColoured = aec.GreenF.Apply(figletStr)
	}
	fmt.Printf(figletColoured)
}

func getGatewayDetails(m map[string]interface{}) (version, sha, commit string) {
	if _, ok := m["orchestration"]; !ok {
		v := m["version"].(map[string]interface{})
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
