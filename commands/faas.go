// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultGateway = "http://localhost:8080"
const defaultNetwork = "func_functions"

// Flags that are to be added to all commands.
var (
	yamlFile string
	regex    string
	filter   string
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

func init() {
	faasCmd.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")
	faasCmd.PersistentFlags().StringVarP(&regex, "regex", "", "", "Regex to match with function names in YAML file")
	faasCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "Wildcard to match with function names in YAML file")

	// Set Bash completion options
	validYAMLFilenames := []string{"yaml", "yml"}
	_ = faasCmd.PersistentFlags().SetAnnotation("yaml", cobra.BashCompFilenameExt, validYAMLFilenames)
}

// Execute TODO
func Execute(customArgs []string) {
	faasCmd.SilenceUsage = true
	faasCmd.SetArgs(customArgs[1:])
	if err := faasCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// faasCmd is the FaaS CLI root command and mimics the legacy client behaviour
// Every other command attached to FaasCmd is a child command to it.
var faasCmd = &cobra.Command{
	Use:   "faas-cli",
	Short: "Manage your OpenFaaS functions from the command line",
	Long: `
Manage your OpenFaaS functions from the command line`,
	Run: runFaas,
}

// runFaas TODO
func runFaas(cmd *cobra.Command, args []string) {
	fmt.Printf(figletStr)
	cmd.Help()
}
