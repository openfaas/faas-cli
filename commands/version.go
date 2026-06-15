// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"os"

	"github.com/alexellis/arkade/pkg/get"
	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/version"
	gatewayTypes "github.com/openfaas/faas/gateway/types"
	"github.com/openfaas/go-sdk/stack"
	"github.com/spf13/cobra"
)

// GitCommit injected at build-time
var (
	shortVersion bool
	warnUpdate   bool
)

type versionInfo struct {
	CLI struct {
		Commit  string `json:"commit"`
		Version string `json:"version"`
	} `json:"cli"`
	Gateway struct {
		URI     string `json:"uri"`
		Version string `json:"version"`
		SHA     string `json:"sha"`
	} `json:"gateway,omitempty"`
	Provider struct {
		Name          string `json:"name"`
		Orchestration string `json:"orchestration"`
		Version       string `json:"version"`
		SHA           string `json:"sha"`
	} `json:"provider,omitempty"`
}

type versionGatewayInfo struct {
	GatewayAddress string
	GatewayInfo    gatewayTypes.GatewayInfo
}

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short-version", false, "Just print Git SHA")
	versionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	versionCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	versionCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yaml file")

	versionCmd.Flags().BoolVar(&warnUpdate, "warn-update", true, "Check for new version and warn about updating")
	versionCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output version as JSON")

	versionCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
	faasCmd.AddCommand(versionCmd)
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version [--short-version] [--gateway GATEWAY_URL] [--json]",
	Short: "Display the clients version information",
	Long: fmt.Sprintf(`The version command returns the current clients version information.

This currently consists of the GitSHA from which the client was built.
- https://github.com/openfaas/faas-cli/tree/%s`, version.GitCommit),
	Example: `  faas-cli version
  faas-cli version --short-version
  faas-cli version --json`,
	RunE: runVersionE,
}

func runVersionE(cmd *cobra.Command, args []string) error {
	if shortVersion {
		fmt.Println(version.BuildVersion())
		return nil
	}

	if jsonOutput {
		info, err := getVersionInfo()
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	printLogo()
	fmt.Printf(`CLI:
  commit:  %s
  version: %s
`, version.GitCommit, version.BuildVersion())
	printServerVersions()

	if warnUpdate {
		version := version.Version
		latest, err := get.FindGitHubRelease("openfaas", "faas-cli")
		if err != nil {
			return fmt.Errorf("unable to find latest version online error: %s", err.Error())
		}

		if version != "" && version != latest {
			fmt.Printf("Your faas-cli version (%s) may be out of date. Version: %s is now available on GitHub.\n", version, latest)
		}
	}

	return nil
}

func getVersionInfo() (versionInfo, error) {
	var info versionInfo
	info.CLI.Commit = version.GitCommit
	info.CLI.Version = version.BuildVersion()

	serverInfo, err := getVersionGatewayInfo()
	if err != nil {
		return info, err
	}

	info.Gateway.URI = serverInfo.GatewayAddress
	info.Gateway.Version = serverInfo.GatewayInfo.Version.Release
	info.Gateway.SHA = serverInfo.GatewayInfo.Version.SHA
	info.Provider.Name = serverInfo.GatewayInfo.Provider.Name
	info.Provider.Orchestration = serverInfo.GatewayInfo.Provider.Orchestration
	info.Provider.Version = serverInfo.GatewayInfo.Provider.Version.Release
	info.Provider.SHA = serverInfo.GatewayInfo.Provider.Version.SHA

	return info, nil
}

func getVersionGatewayInfo() (versionGatewayInfo, error) {
	var info versionGatewayInfo

	var services stack.Services
	var yamlGateway string
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err == nil && parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress := getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))
	info.GatewayAddress = gatewayAddress

	versionTimeout := 5 * time.Second
	cliAuth, err := proxy.NewCLIAuth(token, gatewayAddress)
	if err != nil {
		return info, err
	}
	transport := GetDefaultCLITransport(tlsInsecure, &versionTimeout)
	cliClient, err := proxy.NewClient(cliAuth, gatewayAddress, transport, &versionTimeout)
	if err != nil {
		return info, err
	}

	gatewayInfo, err := cliClient.GetSystemInfo(context.Background())
	if err != nil {
		return info, err
	}

	info.GatewayInfo = gatewayInfo

	return info, nil
}

func printServerVersions() error {
	serverInfo, err := getVersionGatewayInfo()
	if err != nil {
		return err
	}

	printGatewayDetails(serverInfo.GatewayAddress, serverInfo.GatewayInfo.Version.Release, serverInfo.GatewayInfo.Version.SHA)

	fmt.Printf(`
Provider
 name:          %s
 orchestration: %s
 version:       %s 
 sha:           %s
`, serverInfo.GatewayInfo.Provider.Name, serverInfo.GatewayInfo.Provider.Orchestration, serverInfo.GatewayInfo.Provider.Version.Release, serverInfo.GatewayInfo.Provider.Version.SHA)
	return nil
}

func printGatewayDetails(gatewayAddress, version, sha string) {
	fmt.Printf(`
Gateway
 uri:     %s`, gatewayAddress)

	if version != "" {
		fmt.Printf(`
 version: %s
 sha:     %s
`, version, sha)
	}

	fmt.Println()
}

// printLogo prints an ASCII logo, which was generated with figlet
func printLogo() {
	figletColoured := aec.BlueF.Apply(figletStr)
	if runtime.GOOS == "windows" {
		figletColoured = aec.GreenF.Apply(figletStr)
	}
	fmt.Printf("%s", figletColoured)
}

const figletStr = `  ___                   _____           ____
 / _ \ _ __   ___ _ __ |  ___|_ _  __ _/ ___|
| | | | '_ \ / _ \ '_ \| |_ / _` + "`" + ` |/ _` + "`" + ` \___ \
| |_| | |_) |  __/ | | |  _| (_| | (_| |___) |
 \___/| .__/ \___|_| |_|_|  \__,_|\__,_|____/
      |_|

`
