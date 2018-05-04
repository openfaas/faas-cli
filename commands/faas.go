// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"
)

const (
	defaultGateway = "http://127.0.0.1:8080"
	defaultNetwork = ""
	defaultYAML    = "stack.yml"
)

// Flags that are to be added to all commands.
var (
	yamlFile         string
	regex            string
	filter           string
	usingDefaultYaml = false
)

// Flags that are to be added to subset of commands.
var (
	fprocess     string
	functionName string
	network      string
	gateway      string
	handler      string
	image        string
	imagePrefix  string
	language     string
)

var stat = func(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// TODO: remove this workaround once these vars are no longer global
func resetForTest() {
	yamlFile = ""
	regex = ""
	filter = ""
	usingDefaultYaml = false
}

func init() {
	// Setup terminal std
	term.StdStreams()

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
	faasCmd.SilenceErrors = true
	faasCmd.SetArgs(customArgs[1:])

	if err := faasCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}

func checkAndSetDefaultYaml() {
	_, err := stat(defaultYAML)
	// Check if there is a default yaml file and set it
	if len(yamlFile) == 0 && err == nil {
		yamlFile = defaultYAML
		usingDefaultYaml = true
	}
}

// faasCmd is the FaaS CLI root command and mimics the legacy client behaviour
// Every other command attached to FaasCmd is a child command to it.
var faasCmd = &cobra.Command{
	Use:   "faas-cli",
	Short: "Manage your OpenFaaS functions from the command line",
	Long: `
Manage your OpenFaaS functions from the command line`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		checkAndSetDefaultYaml()
	},
	Run: runFaas,
}

// runFaas TODO
func runFaas(cmd *cobra.Command, args []string) {
	printFiglet()
	cmd.Help()
}
