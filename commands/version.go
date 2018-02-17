// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/version"
	"github.com/spf13/cobra"
)

const (
	githubApiReleases = "https://api.github.com/repos/openfaas/faas-cli/releases"
)

// GitCommit injected at build-time
var (
	shortVersion bool
)

type githubApiRelease struct {
	PreRelease  bool      `json:"prerelease"`
	Tag         string    `json:"tag_name"`
	HtmlUrl     string    `json:"html_url"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
}

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short-version", false, "Just print the version")

	faasCmd.AddCommand(versionCmd)
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version [--short-version]",
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
		fmt.Printf("Commit: %s\n", version.GitCommit)
		fmt.Printf("Version: %s\n", version.BuildVersion())

		if !version.IsDev() {
			latestVersion, _ := getLatestVersion()
			if latestVersion != nil && version.CompareVersion(version.BuildVersion(), latestVersion.Tag) == 1 {
				fmt.Printf(aec.YellowF.Apply("Newer version available: %s\n"), latestVersion.Tag)
			}
		}
	}
}

func printFiglet() {
	figletColoured := aec.BlueF.Apply(figletStr)
	if runtime.GOOS == "windows" {
		figletColoured = aec.GreenF.Apply(figletStr)
	}
	fmt.Printf(figletColoured)
}

func getLatestVersion() (*githubApiRelease, error) {
	timeout := 1 * time.Second
	client := proxy.MakeHTTPClient(&timeout)

	if resp, err := client.Get(githubApiReleases); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		if response, err := ioutil.ReadAll(resp.Body); err != nil {
			return nil, err
		} else {
			var data []githubApiRelease
			if err := json.Unmarshal(response, &data); err != nil {
				return nil, err
			}

			for _, r := range data {
				if r.PreRelease == false {
					return &r, nil
				}
			}

			return nil, errors.New("could not determine the latest version")
		}
	}
}

const figletStr = `  ___                   _____           ____
 / _ \ _ __   ___ _ __ |  ___|_ _  __ _/ ___|
| | | | '_ \ / _ \ '_ \| |_ / _` + "`" + ` |/ _` + "`" + ` \___ \
| |_| | |_) |  __/ | | |  _| (_| | (_| |___) |
 \___/| .__/ \___|_| |_|_|  \__,_|\__,_|____/
      |_|

`
