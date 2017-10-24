// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/analytics"
	"github.com/spf13/cobra"
)

const (
	analyticsTimeout = time.Millisecond * 300
	defaultGateway   = "http://localhost:8080"
	defaultNetwork   = "func_functions"
	defaultYAML      = "stack.yml"
)

var analyticsCh chan int

// Flags that are to be added to all commands.
var (
	yamlFile         string
	regex            string
	filter           string
	disableAnalytics bool
)

// Flags that are to be added to subset of commands.
var (
	fprocess     string
	functionName string
	network      string
	gateway      string
	handler      string
	image        string
	language     string
)

var stat = func(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

func init() {
	faasCmd.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")
	faasCmd.PersistentFlags().StringVarP(&regex, "regex", "", "", "Regex to match with function names in YAML file")
	faasCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "Wildcard to match with function names in YAML file")

	// Set Bash completion options
	validYAMLFilenames := []string{"yaml", "yml"}
	_ = faasCmd.PersistentFlags().SetAnnotation("yaml", cobra.BashCompFilenameExt, validYAMLFilenames)

	analyticsCh = make(chan int)
}

// Execute TODO
func Execute(customArgs []string) {
	checkAndSetDefaultYaml()

	faasCmd.SilenceUsage = true
	faasCmd.SilenceErrors = true
	faasCmd.SetArgs(customArgs[1:])
	if err := faasCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}

	if analytics.Disabled() {
		return
	}

	// Block on the submission of an analytics event or timeout expiration, this
	// is to allow sufficient time for the event submission goroutine to complete
	// when the Gateway responds extremely quickly.
	select {
	case <-analyticsCh:
	case <-time.After(analyticsTimeout):
	}
}

func checkAndSetDefaultYaml() {
	// Check if there is a default yaml file and set it
	if _, err := stat(defaultYAML); err == nil {
		yamlFile = defaultYAML
	}
}

// faasCmd is the FaaS CLI root command and mimics the legacy client behaviour
// Every other command attached to FaasCmd is a child command to it.
var faasCmd = &cobra.Command{
	Use:   "faas-cli",
	Short: "Manage your OpenFaaS functions from the command line",
	Long: `
Manage your OpenFaaS functions from the command line`,
	PreRun: func(cmd *cobra.Command, args []string) {
		analytics.Event("root", "", analyticsCh)
	},
	Run: runFaas,
}

// runFaas TODO
func runFaas(cmd *cobra.Command, args []string) {
	fmt.Printf(figletStr)
	cmd.Help()
}
